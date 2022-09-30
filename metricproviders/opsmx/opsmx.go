package opsmx

import (
	"context"
	"encoding/json"

	"math"
	"path"

	"net/http"
	"net/url"

	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"time"

	"errors"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/utils/defaults"
	metricutil "github.com/argoproj/argo-rollouts/utils/metric"
	timeutil "github.com/argoproj/argo-rollouts/utils/time"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ProviderType                            = "opsmx"
	v5configIdLookupURLFormat               = `/autopilot/api/v5/registerCanary`
	scoreUrlFormat                          = `/autopilot/canaries/`
	reportUrlFormat                         = `ui/application/deploymentverification/`
	resumeAfter                             = 3 * time.Second
	httpConnectionTimeout     time.Duration = 15 * time.Second
	defaultConfigMapName                    = "opsmx-profile"
	defaultcdIntegration                    = "argorollouts"
	cdIntegrationArgoCD                     = "argocd"
)

type Provider struct {
	logCtx        log.Entry
	kubeclientset kubernetes.Interface
	client        http.Client
}

type jobPayload struct {
	Application       string              `json:"application"`
	SourceName        string              `json:"sourceName"`
	SourceType        string              `json:"sourceType"`
	CanaryConfig      canaryConfig        `json:"canaryConfig"`
	CanaryDeployments []canaryDeployments `json:"canaryDeployments"`
}

type canaryConfig struct {
	LifetimeMinutes          string                   `json:"lifetimeMinutes"`
	LookBackType             string                   `json:"lookbackType,omitempty"`
	IntervalTime             string                   `json:"interval,omitempty"`
	CanaryHealthCheckHandler canaryHealthCheckHandler `json:"canaryHealthCheckHandler"`
	CanarySuccessCriteria    canarySuccessCriteria    `json:"canarySuccessCriteria"`
}

type canaryHealthCheckHandler struct {
	MinimumCanaryResultScore string `json:"minimumCanaryResultScore"`
}

type canarySuccessCriteria struct {
	CanaryResultScore string `json:"canaryResultScore"`
}

type canaryDeployments struct {
	CanaryStartTimeMs   string     `json:"canaryStartTimeMs"`
	BaselineStartTimeMs string     `json:"baselineStartTimeMs"`
	Canary              *logMetric `json:"canary,omitempty"`
	Baseline            *logMetric `json:"baseline,omitempty"`
}
type logMetric struct {
	Log    map[string]map[string]string `json:"log,omitempty"`
	Metric map[string]map[string]string `json:"metric,omitempty"`
}

// Type indicates provider is a OPSMX provider
func (*Provider) Type() string {
	return ProviderType
}

// GetMetadata returns any additional metadata which needs to be stored & displayed as part of the metrics result.
func (*Provider) GetMetadata(metric v1alpha1.Metric) map[string]string {
	return nil
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func urlJoiner(gateUrl string, paths ...string) (string, error) {
	u, err := url.Parse(gateUrl)
	if err != nil {
		return "", err
	}
	for _, p := range paths {
		u.Path = path.Join(u.Path, p)

	}
	return u.String(), nil
}

func makeRequest(client http.Client, requestType string, url string, body string, user string) ([]byte, error) {
	reqBody := strings.NewReader(body)
	req, err := http.NewRequest(
		requestType,
		url,
		reqBody,
	)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("x-spinnaker-user", user)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Infof("In error")
		return []byte{}, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return data, err
}

// Check few conditions pre-analysis
func basicChecks(metric v1alpha1.Metric) error {
	if metric.Provider.OPSMX.Threshold.Pass <= metric.Provider.OPSMX.Threshold.Marginal {
		return errors.New("pass score cannot be less than marginal score")
	}
	if metric.Provider.OPSMX.LifetimeMinutes == "" && metric.Provider.OPSMX.EndTime == "" {
		return errors.New("either provide lifetimeMinutes or end time")
	}
	if metric.Provider.OPSMX.CanaryStartTime != metric.Provider.OPSMX.BaselineStartTime && metric.Provider.OPSMX.LifetimeMinutes == "" {
		return errors.New("both start time should be kept same in case of using end time argument")
	}
	if metric.Provider.OPSMX.LifetimeMinutes != "" {
		lifetimeMinutes, err := strconv.Atoi(metric.Provider.OPSMX.LifetimeMinutes)
		if err != nil {
			return errors.New("lifetime minutes should be in integer format")
		}
		if lifetimeMinutes < 3 {
			return errors.New("lifetime minutes cannot be less than 3 minutes")
		}
	}
	if metric.Provider.OPSMX.IntervalTime != "" {
		intervalTime, err := strconv.Atoi(metric.Provider.OPSMX.IntervalTime)
		if err != nil {
			return errors.New("interval time should be in integer format")
		}
		if intervalTime < 3 {
			return errors.New("interval time cannot be less than 3 minutes")
		}
	}
	return nil
}

// Return epoch values of the specific time provided along with lifetimeMinutes for the Run
func getTimeVariables(baselineTime string, canaryTime string, endTime string, lifetimeMinutes string) (string, string, string, error) {

	var canaryStartTime string
	var baselineStartTime string
	tm := time.Now()

	if canaryTime == "" {
		canaryStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
	} else {
		tsStart, err := time.Parse(time.RFC3339, canaryTime)
		if err != nil {
			return "", "", "", err
		}
		canaryStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond))
	}

	if baselineTime == "" {
		baselineStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
	} else {
		tsStart, err := time.Parse(time.RFC3339, baselineTime)
		if err != nil {
			return "", "", "", err
		}
		baselineStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond))
	}

	//If lifetimeMinutes not given calculate using endTime
	if lifetimeMinutes == "" {
		tsEnd, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return "", "", "", err
		}
		if canaryTime != "" && canaryTime > endTime {
			err := errors.New("start time cannot be greater than end time")
			return "", "", "", err
		}
		tsStart := tm
		if canaryTime != "" {
			tsStart, _ = time.Parse(time.RFC3339, canaryTime)
		}
		tsDifference := tsEnd.Sub(tsStart)
		min, _ := time.ParseDuration(tsDifference.String())
		lifetimeMinutes = fmt.Sprintf("%v", roundFloat(min.Minutes(), 0))
	}
	return canaryStartTime, baselineStartTime, lifetimeMinutes, nil
}

func getDataConfigMap(metric v1alpha1.Metric, kubeclientset kubernetes.Interface, isRun bool) (map[string]string, error) {
	//TODO - Refactor
	ns := defaults.Namespace()
	cmData := map[string]string{}
	configMapName := defaultConfigMapName
	if metric.Provider.OPSMX.Profile != "" {
		configMapName = metric.Provider.OPSMX.Profile
	}
	configmap, err := kubeclientset.CoreV1().Secrets(ns).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	gateUrl := metric.Provider.OPSMX.GateUrl
	if gateUrl == "" {
		configmapGateurl, ok := configmap.Data["gate-url"]
		if !ok {
			err := errors.New("the gate-url is not specified both in the template and in the secret")
			return nil, err
		}
		gateUrl = string(configmapGateurl)
	}
	cmData["gateUrl"] = gateUrl

	user := metric.Provider.OPSMX.User
	if user == "" {
		configmapUser, ok := configmap.Data["user"]
		if !ok {
			err := errors.New("the user is not specified both in the template and in the secret")
			return nil, err
		}
		user = string(configmapUser)
	}
	cmData["user"] = user

	if isRun == false {
		return cmData, nil
	}

	//TODO - Check for yaml bool types
	cdIntegration := defaultcdIntegration
	configMapCdIntegration, ok := configmap.Data["cd-integration"]
	if ok {
		if string(configMapCdIntegration) == "true" {
			cdIntegration = cdIntegrationArgoCD
		}
		if string(configMapCdIntegration) != "true" && string(configMapCdIntegration) != "false" {
			err := errors.New("cd-integration should be either true or false")
			return nil, err
		}
	}
	cmData["cdIntegration"] = cdIntegration

	configmapSourceName, ok := configmap.Data["source-name"]
	if !ok {
		err := errors.New("source-name is not specified in the secret")
		return nil, err
	}
	cmData["sourceName"] = string(configmapSourceName)

	return cmData, nil
}

// Run queries opsmx for the metric
func (p *Provider) Run(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := timeutil.MetaNow()
	newMeasurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}
	log.Infof("Before getDataConfigMap")
	configMapData, err := getDataConfigMap(metric, p.kubeclientset, true)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	log.Infof("%s", configMapData)
	log.Infof("After getDataConfigMap")

	//develop Canary Register Url
	canaryurl, err := urlJoiner(configMapData["gateUrl"], v5configIdLookupURLFormat)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	//Run basicChecks
	if err := basicChecks(metric); err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	//Get the epochs for Time variables and the lifetimeMinutes
	canaryStartTime, baselineStartTime, lifetimeMinutes, err := getTimeVariables(metric.Provider.OPSMX.BaselineStartTime, metric.Provider.OPSMX.CanaryStartTime, metric.Provider.OPSMX.EndTime, metric.Provider.OPSMX.LifetimeMinutes)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	//Generate the payload
	payload := jobPayload{
		Application: metric.Provider.OPSMX.Application,
		SourceName:  configMapData["sourceName"],
		SourceType:  configMapData["cdIntegration"],
		CanaryConfig: canaryConfig{
			LifetimeMinutes: lifetimeMinutes,
			LookBackType:    metric.Provider.OPSMX.LookBackType,
			IntervalTime:    metric.Provider.OPSMX.IntervalTime,
			CanaryHealthCheckHandler: canaryHealthCheckHandler{
				MinimumCanaryResultScore: fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Marginal),
			},
			CanarySuccessCriteria: canarySuccessCriteria{
				CanaryResultScore: fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Pass),
			},
		},
		CanaryDeployments: []canaryDeployments{},
	}

	if metric.Provider.OPSMX.Services != nil || len(metric.Provider.OPSMX.Services) != 0 {
		deployment := canaryDeployments{
			BaselineStartTimeMs: baselineStartTime,
			CanaryStartTimeMs:   canaryStartTime,
			Baseline: &logMetric{
				Log:    map[string]map[string]string{},
				Metric: map[string]map[string]string{},
			},
			Canary: &logMetric{
				Log:    map[string]map[string]string{},
				Metric: map[string]map[string]string{},
			},
		}
		for i, item := range metric.Provider.OPSMX.Services {
			valid := false
			serviceName := fmt.Sprintf("service%d", i+1)
			if item.ServiceName != "" {
				serviceName = item.ServiceName
			}
			gateName := fmt.Sprintf("gate%d", i+1)
			//For Log Analysis is to be added in analysis-run
			if item.LogScopeVariables != "" {
				//Check if no baseline or canary
				if item.BaselineLogScope == "" || item.CanaryLogScope == "" {
					err := errors.New("missing baseline/canary for log analysis")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				//Check if the number of placeholders provided dont match
				if len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.BaselineLogScope, ",")) || len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.CanaryLogScope, ",")) {
					err := errors.New("mismatch in number of log scope variables and baseline/canary log scope")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				if item.LogTemplateName == "" && metric.Provider.OPSMX.GlobalLogTemplate == "" {
					err := errors.New("provide either a service specific log template or global log template")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				//Add mandatory field for baseline
				deployment.Baseline.Log[serviceName] = map[string]string{
					item.LogScopeVariables: item.BaselineLogScope,
					"serviceGate":          gateName,
				}
				//Add mandatory field for canary
				deployment.Canary.Log[serviceName] = map[string]string{
					item.LogScopeVariables: item.CanaryLogScope,
					"serviceGate":          gateName,
				}

				//Add service specific templateName
				if item.LogTemplateName != "" {
					deployment.Baseline.Log[serviceName]["template"] = item.LogTemplateName
					deployment.Canary.Log[serviceName]["template"] = item.LogTemplateName
				} else {
					deployment.Baseline.Log[serviceName]["template"] = metric.Provider.OPSMX.GlobalLogTemplate
					deployment.Canary.Log[serviceName]["template"] = metric.Provider.OPSMX.GlobalLogTemplate
				}

				//Add non-mandatory field of Templateversion if provided
				if item.LogTemplateVersion != "" {
					deployment.Baseline.Log[serviceName]["templateVersion"] = item.LogTemplateVersion
					deployment.Canary.Log[serviceName]["templateVersion"] = item.LogTemplateVersion
				}
				valid = true
			}

			//For metric analysis is to be added in analysis-run
			if item.MetricScopeVariables != "" {
				//Check if no baseline or canary
				if item.BaselineMetricScope == "" || item.CanaryMetricScope == "" {
					err := errors.New("missing baseline/canary for metric analysis")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				//Check if the number of placeholders provided dont match
				if len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.BaselineMetricScope, ",")) || len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.CanaryMetricScope, ",")) {
					err := errors.New("mismatch in number of metric scope variables and baseline/canary metric scope")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				if item.MetricTemplateName == "" && metric.Provider.OPSMX.GlobalMetricTemplate == "" {
					err := errors.New("provide either a service specific metric template or global metric template")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}
				//Add mandatory field for baseline
				deployment.Baseline.Metric[serviceName] = map[string]string{
					item.MetricScopeVariables: item.BaselineMetricScope,
					"serviceGate":             gateName,
				}
				//Add mandatory field for canary
				deployment.Canary.Metric[serviceName] = map[string]string{
					item.MetricScopeVariables: item.CanaryMetricScope,
					"serviceGate":             gateName,
				}
				//Add templateName
				if item.MetricTemplateName != "" {
					deployment.Baseline.Metric[serviceName]["template"] = item.MetricTemplateName
					deployment.Canary.Metric[serviceName]["template"] = item.MetricTemplateName
				} else {
					deployment.Baseline.Metric[serviceName]["template"] = metric.Provider.OPSMX.GlobalMetricTemplate
					deployment.Canary.Metric[serviceName]["template"] = metric.Provider.OPSMX.GlobalMetricTemplate
				}

				//Add non-mandatory field of Template Version if provided
				if item.MetricTemplateVersion != "" {
					deployment.Baseline.Metric[serviceName]["templateVersion"] = item.MetricTemplateVersion
					deployment.Canary.Metric[serviceName]["templateVersion"] = item.MetricTemplateVersion
				}
				valid = true

			}
			//Check if no logs or metrics were provided
			if !valid {
				err := errors.New("at least one of log or metric context must be included")
				return metricutil.MarkMeasurementError(newMeasurement, err)
			}
		}
		payload.CanaryDeployments = append(payload.CanaryDeployments, deployment)
	} else {
		//Check if no services were provided
		err = errors.New("no services provided")
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	buffer, err := json.Marshal(payload)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	log.Infof("Before make request")
	data, err := makeRequest(p.client, "POST", canaryurl, string(buffer), configMapData["user"])
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	//Struct to record canary Response
	type canaryResponse struct {
		Error    string      `json:"error,omitempty"`
		Message  string      `json:"message,omitempty"`
		CanaryId json.Number `json:"canaryId,omitempty"`
	}
	var canary canaryResponse

	json.Unmarshal(data, &canary)

	if canary.Error != "" {
		errMessage := fmt.Sprintf("Error: %s\nMessage: %s", canary.Error, canary.Message)
		err := errors.New(errMessage)
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	//Develop the Report URL
	stringifiedCanaryId := string(canary.CanaryId)
	reportUrl, err := urlJoiner(configMapData["gateUrl"], reportUrlFormat, metric.Provider.OPSMX.Application, stringifiedCanaryId)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	mapMetadata := make(map[string]string)
	mapMetadata["Group"] = run.Name
	mapMetadata["payload"] = string(buffer)
	mapMetadata["gateUrl"] = configMapData["gateUrl"]
	mapMetadata["dataPassedFromSecretsFunc"] = fmt.Sprintf("%s", configMapData)
	mapMetadata["canaryId"] = stringifiedCanaryId
	mapMetadata["reportUrl"] = fmt.Sprintf("Report Url: %s", reportUrl)
	resumeTime := metav1.NewTime(timeutil.Now().Add(resumeAfter))
	newMeasurement.Metadata = mapMetadata
	newMeasurement.ResumeAt = &resumeTime
	newMeasurement.Phase = v1alpha1.AnalysisPhaseRunning
	return newMeasurement
}

// Evaluate canaryScore and accordingly set the AnalysisPhase
func evaluateResult(score int, pass int, marginal int) v1alpha1.AnalysisPhase {
	if score >= pass {
		return v1alpha1.AnalysisPhaseSuccessful
	}
	if score < pass && score >= marginal {
		return v1alpha1.AnalysisPhaseInconclusive
	}
	return v1alpha1.AnalysisPhaseFailed
}

// Extract the canaryScore and evaluateResult
func processResume(data []byte, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	var (
		canaryScore string
		result      map[string]interface{}
		finalScore  map[string]interface{}
	)

	if !json.Valid(data) {
		err := errors.New("invalid Response")
		return metricutil.MarkMeasurementError(measurement, err)
	}

	json.Unmarshal(data, &result)
	jsonBytes, _ := json.MarshalIndent(result["canaryResult"], "", "   ")
	json.Unmarshal(jsonBytes, &finalScore)
	if finalScore["overallScore"] == nil {
		canaryScore = "0"
	} else {
		canaryScore = fmt.Sprintf("%v", finalScore["overallScore"])
	}
	score, _ := strconv.Atoi(canaryScore)
	measurement.Value = canaryScore
	measurement.Phase = evaluateResult(score, int(metric.Provider.OPSMX.Threshold.Pass), int(metric.Provider.OPSMX.Threshold.Marginal))
	return measurement
}

// Resume the in-progress measurement
func (p *Provider) Resume(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	configMapData, err := getDataConfigMap(metric, p.kubeclientset, false)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}

	canaryId := measurement.Metadata["canaryId"]
	scoreURL, err := urlJoiner(configMapData["gateUrl"], scoreUrlFormat, canaryId)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}

	data, err := makeRequest(p.client, "GET", scoreURL, "", configMapData["user"])
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}
	var status map[string]interface{}
	json.Unmarshal(data, &status)
	a, _ := json.MarshalIndent(status["status"], "", "   ")
	json.Unmarshal(a, &status)
	//if the status is Running, resume analysis after delay
	if status["status"] == "RUNNING" {
		resumeTime := metav1.NewTime(timeutil.Now().Add(resumeAfter))
		measurement.ResumeAt = &resumeTime
		measurement.Phase = v1alpha1.AnalysisPhaseRunning
		return measurement
	}
	//if run is cancelled mid-run
	if status["status"] == "CANCELLED" {
		measurement.Phase = v1alpha1.AnalysisPhaseFailed
		measurement.Message = "Analysis Cancelled"
	} else {
		//POST-Run process
		measurement = processResume(data, metric, measurement)
	}
	finishTime := timeutil.MetaNow()
	measurement.FinishedAt = &finishTime
	return measurement
}

// Terminate should not be used the OPSMX provider since all the work should occur in the Run method
func (p *Provider) Terminate(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	p.logCtx.Warn("OPSMX provider should not execute the Terminate method")
	return measurement
}

// GarbageCollect is a no-op for the OPSMX provider
func (p *Provider) GarbageCollect(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, limit int) error {
	return nil
}

func NewOPSMXProvider(logCtx log.Entry, kubeclientset kubernetes.Interface, client http.Client) *Provider {
	return &Provider{
		logCtx:        logCtx,
		kubeclientset: kubeclientset,
		client:        client,
	}
}

func NewHttpClient() http.Client {
	c := http.Client{
		Timeout: httpConnectionTimeout,
	}
	return c
}
