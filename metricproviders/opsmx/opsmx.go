package opsmx

import (
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
	metricutil "github.com/argoproj/argo-rollouts/utils/metric"
	timeutil "github.com/argoproj/argo-rollouts/utils/time"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ProviderType                          = "opsmx"
	configIdLookupURLFormat               = `/autopilot/api/v3/registerCanary`
	scoreUrlFormat                        = `/autopilot/canaries/`
	reportUrlFormat                       = `ui/application/deploymentverification/`
	resumeAfter                           = 3 * time.Second
	httpConnectionTimeout   time.Duration = 15 * time.Second
)

type Provider struct {
	logCtx log.Entry
	client http.Client
}

type jobPayload struct {
	Application       string              `json:"application"`
	CanaryConfig      canaryConfig        `json:"canaryConfig"`
	CanaryDeployments []canaryDeployments `json:"canaryDeployments"`
}

type canaryConfig struct {
	LifetimeHours            string                   `json:"lifetimeHours"`
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
	CanaryStartTimeMs   string    `json:"canaryStartTimeMs"`
	BaselineStartTimeMs string    `json:"baselineStartTimeMs"`
	Canary              logMetric `json:"canary,omitempty"`
	Baseline            logMetric `json:"baseline,omitempty"`
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

func makeRequest(client http.Client, requestType string, url string, body string, user string) ([]byte, error) {
	reqBody := strings.NewReader(body)
	// create a request object
	req, err := http.NewRequest(
		requestType,
		url,
		reqBody,
	)
	if err != nil {
		return []byte{}, err
	}

	// add a request header
	req.Header.Set("x-spinnaker-user", user)
	req.Header.Set("Content-Type", "application/json")

	// send an HTTP using `req` object
	res, err := client.Do(req)
	// check for response error
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return data, err
}

func basicChecks(metric v1alpha1.Metric) error {
	if metric.Provider.OPSMX.CanaryStartTime == "" && metric.Provider.OPSMX.BaselineStartTime == "" && metric.Provider.OPSMX.LifetimeHours == "" {
		return errors.New("either provide lifetimehours or start time")
	}
	if metric.Provider.OPSMX.Threshold.Pass <= metric.Provider.OPSMX.Threshold.Marginal {
		return errors.New("pass score cannot be less than marginal score")
	}
	if metric.Provider.OPSMX.LifetimeHours == "" && metric.Provider.OPSMX.EndTime == "" {
		return errors.New("either provide lifetimehours or end time")
	}
	return nil
}

func getTimeVariables(baselineTime string, canaryTime string, endTime string, lifetimeHours string) (string, string, string, error) {

	var canaryStartTime string
	var baselineStartTime string

	if (canaryTime == "" && baselineTime != "") || (canaryTime != "" && baselineTime == "") {
		if canaryTime == "" {
			canaryTime = baselineTime
		} else {
			baselineTime = canaryTime
		}
	}

	if canaryTime == "" && baselineTime == "" {
		tm := time.Now()
		canaryStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
		baselineStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
	} else {
		tsStart, err := time.Parse(time.RFC3339, canaryTime) //make a time object for canary start time
		if err != nil {
			return "", "", "", err
		}
		canaryStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond)) //convert ISO to epoch

		tsStart, err = time.Parse(time.RFC3339, baselineTime) //make a time object for baseline start time
		if err != nil {
			return "", "", "", err
		}
		baselineStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond)) //convert ISO to epoch
	}

	//If lifetimeHours not given
	if lifetimeHours == "" {
		tsStart, _ := time.Parse(time.RFC3339, canaryTime)
		tsEnd, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return "", "", "", err
		}
		if canaryTime > endTime {
			err := errors.New("start time cannot be greater than end time")
			return "", "", "", err
		}
		tsDifference := tsEnd.Sub(tsStart)
		min, _ := time.ParseDuration(tsDifference.String())
		lifetimeHours = fmt.Sprintf("%v", roundFloat(min.Minutes()/60, 1))
	}

	return canaryStartTime, baselineStartTime, lifetimeHours, nil
}

// Run queries opsmx for the metric
func (p *Provider) Run(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := timeutil.MetaNow()
	newMeasurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}

	urlCanary, err := url.Parse(metric.Provider.OPSMX.GateUrl)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	urlCanary.Path = path.Join(urlCanary.Path, configIdLookupURLFormat)
	canaryurl := urlCanary.String()

	if err := basicChecks(metric); err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	canaryStartTime, baselineStartTime, lifetimeHours, err := getTimeVariables(metric.Provider.OPSMX.BaselineStartTime, metric.Provider.OPSMX.CanaryStartTime, metric.Provider.OPSMX.EndTime, metric.Provider.OPSMX.LifetimeHours)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	payload := jobPayload{
		Application: metric.Provider.OPSMX.Application,
		CanaryConfig: canaryConfig{
			LifetimeHours: lifetimeHours,
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
			Baseline: logMetric{
				Log:    map[string]map[string]string{},
				Metric: map[string]map[string]string{},
			},
			Canary: logMetric{
				Log:    map[string]map[string]string{},
				Metric: map[string]map[string]string{},
			},
		}
		for _, item := range metric.Provider.OPSMX.Services {
			valid := false
			if item.LogScopeVariables != "" {
				if len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.BaselineLogScope, ",")) || len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.CanaryLogScope, ",")) {
					err := errors.New("mismatch in amount of log scope variables and baseline/canary log scope")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}

				deployment.Baseline.Log[item.ServiceName] = map[string]string{
					item.LogScopeVariables: item.BaselineLogScope,
					"serviceGate":          item.GateName,
				}
				deployment.Canary.Log[item.ServiceName] = map[string]string{
					item.LogScopeVariables: item.CanaryLogScope,
					"serviceGate":          item.GateName,
				}
				valid = true
			}

			if item.MetricScopeVariables != "" {
				if len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.BaselineMetricScope, ",")) || len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.CanaryMetricScope, ",")) {
					err := errors.New("mismatch in amount of log scope variables and baseline/canary log scope")
					return metricutil.MarkMeasurementError(newMeasurement, err)
				}

				deployment.Baseline.Metric[item.ServiceName] = map[string]string{
					item.MetricScopeVariables: item.BaselineMetricScope,
					"serviceGate":             item.GateName,
				}
				deployment.Canary.Metric[item.ServiceName] = map[string]string{
					item.MetricScopeVariables: item.CanaryMetricScope,
					"serviceGate":             item.GateName,
				}
				valid = true

			}
			if !valid {
				err := errors.New("at least one of log or metric context must be included")
				return metricutil.MarkMeasurementError(newMeasurement, err)
			}
		}
		payload.CanaryDeployments = append(payload.CanaryDeployments, deployment)
	} else {
		internal := canaryDeployments{
			BaselineStartTimeMs: baselineStartTime,
			CanaryStartTimeMs:   canaryStartTime,
		}
		payload.CanaryDeployments = append(payload.CanaryDeployments, internal)
	}
	buffer, err := json.Marshal(payload)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	// create a request object
	data, err := makeRequest(p.client, "POST", canaryurl, string(buffer), metric.Provider.OPSMX.User)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	type canaryResponse struct {
		Error    string `json:"error,omitempty"`
		CanaryId string `json:"canaryId,omitempty"`
	}
	var canary canaryResponse

	json.Unmarshal(data, &canary)
	if canary.Error != "" {
		err := errors.New(canary.Error)
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	urlReport, err := url.Parse(metric.Provider.OPSMX.GateUrl)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	urlReport.Path = path.Join(urlReport.Path, reportUrlFormat, metric.Provider.OPSMX.Application, canary.CanaryId)
	reportUrl := urlReport.String()

	//creating a map to return the reporturl and associated data
	mapMetadata := make(map[string]string)
	mapMetadata["canaryId"] = canary.CanaryId
	mapMetadata["reportUrl"] = fmt.Sprintf("Report Url: %s", reportUrl)
	resumeTime := metav1.NewTime(timeutil.Now().Add(resumeAfter))
	newMeasurement.Metadata = mapMetadata
	newMeasurement.ResumeAt = &resumeTime
	newMeasurement.Phase = v1alpha1.AnalysisPhaseRunning
	return newMeasurement
}

func evaluateResult(score int, pass int, marginal int) v1alpha1.AnalysisPhase {
	if score >= pass {
		return v1alpha1.AnalysisPhaseSuccessful
	}
	if score < pass && score >= marginal {
		return v1alpha1.AnalysisPhaseInconclusive
	}
	return v1alpha1.AnalysisPhaseFailed
}

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

func (p *Provider) Resume(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	canaryId := measurement.Metadata["canaryId"]
	urlScore, err := url.Parse(metric.Provider.OPSMX.GateUrl)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}
	urlScore.Path = path.Join(urlScore.Path, scoreUrlFormat, canaryId)
	scoreURL := urlScore.String()

	data, err := makeRequest(p.client, "GET", scoreURL, "", metric.Provider.OPSMX.User)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}
	var status map[string]interface{}
	json.Unmarshal(data, &status)
	a, _ := json.MarshalIndent(status["status"], "", "   ")
	json.Unmarshal(a, &status)
	//return the measurement if the status is Running, to be resumed at resumeTime
	if status["status"] == "RUNNING" {
		resumeTime := metav1.NewTime(timeutil.Now().Add(resumeAfter))
		measurement.ResumeAt = &resumeTime
		measurement.Phase = v1alpha1.AnalysisPhaseRunning
		return measurement
	}
	measurement = processResume(data, metric, measurement)
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

func NewOPSMXProvider(logCtx log.Entry, client http.Client) *Provider {
	return &Provider{
		logCtx: logCtx,
		client: client,
	}
}

func NewHttpClient() http.Client {
	c := http.Client{
		Timeout: httpConnectionTimeout,
	}
	return c
}
