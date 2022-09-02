package opsmx

import (
	"encoding/json"

	"math"

	"net/http"

	"fmt"
	"io/ioutil"
	"regexp"
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
	configIdLookupURLFormat               = `%s/autopilot/api/v3/registerCanary`
	scoreUrlFormat                        = `%s/autopilot/canaries/%s`
	reportUrlFormat                       = `%sui/application/deploymentverification/%s/%s`
	resumeAfter                           = 3 * time.Second
	httpConnectionTimeout   time.Duration = 15 * time.Second
	internalFormat                        = `%s":{serviceGate:%s,%s:%s}`
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
	CanaryStartTimeMs   string     `json:"canaryStartTimeMs"`
	BaselineStartTimeMs string     `json:"baselineStartTimeMs"`
	Canary              *logMetric `json:"canary,omitempty"`
	Baseline            *logMetric `json:"baseline,omitempty"`
}

type logMetric struct {
	Log    string `json:"log,omitempty"`
	Metric string `json:"metric,omitempty"`
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

func makeRequest(requestType string, url string, body string, user string) ([]byte, error) {
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
	res, err := http.DefaultClient.Do(req)
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

func getServicePayload(metric v1alpha1.Metric) (string, string, string, string, error) {

	var (
		baselinelog    string
		canarylog      string
		baselinemetric string
		canarymetric   string
	)

	for _, item := range metric.Provider.OPSMX.Services {
		if item.LogScopeVariables != "" {
			if len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.BaselineLogScope, ",")) || len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.CanaryLogScope, ",")) {
				err := errors.New("mismatch in amount of log scope variables and baseline/canary log scope")
				return "", "", "", "", err
			}
			if baselinelog == "" {
				baselinelog = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.BaselineLogScope)
				canarylog = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.CanaryLogScope)
			} else {
				temp := fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.BaselineLogScope)
				baselinelog = fmt.Sprintf("%s,%s", baselinelog, temp)
				temp = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.CanaryLogScope)
				canarylog = fmt.Sprintf("%s,%s", canarylog, temp)
			}
		}
		if item.MetricScopeVariables != "" {
			if len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.BaselineMetricScope, ",")) || len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.CanaryMetricScope, ",")) {
				err := errors.New("mismatch in amount of metric scope variables and baseline/canary metric scope")
				return "", "", "", "", err
			}
			if baselinemetric == "" {
				baselinemetric = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.BaselineMetricScope)
				canarymetric = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.CanaryMetricScope)
			} else {
				temp := fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.BaselineMetricScope)
				baselinemetric = fmt.Sprintf("%s,%s", baselinemetric, temp)
				temp = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.CanaryMetricScope)
				canarymetric = fmt.Sprintf("%s,%s", canarymetric, temp)
			}
		}
	}
	if baselinelog == "" && baselinemetric == "" {
		err := errors.New("either provide log or metric arguments")
		return "", "", "", "", err
	}
	return baselinelog, canarylog, baselinemetric, canarymetric, nil
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

func getTimeVariables(BaselineTime string, CanaryTime string, EndTime string, lifetimeHours string) (string, string, string, error) {

	var canaryStartTime string
	var baselineStartTime string

	if (CanaryTime == "" && BaselineTime != "") || (CanaryTime != "" && BaselineTime == "") {
		if CanaryTime == "" {
			CanaryTime = BaselineTime
		} else {
			BaselineTime = CanaryTime
		}
	}
	if CanaryTime == "" && BaselineTime == "" {
		tm := time.Now()
		canaryStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
		baselineStartTime = fmt.Sprintf("%d", tm.UnixNano()/int64(time.Millisecond))
	} else {
		tsStart, err := time.Parse(time.RFC3339, CanaryTime) //make a time object for canary start time
		if err != nil {
			return "", "", "", err
		}
		canaryStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond)) //convert ISO to epoch

		tsStart, err = time.Parse(time.RFC3339, BaselineTime) //make a time object for baseline start time
		if err != nil {
			return "", "", "", err
		}
		baselineStartTime = fmt.Sprintf("%d", tsStart.UnixNano()/int64(time.Millisecond)) //convert ISO to epoch
	}

	//If lifetimeHours not given
	if lifetimeHours == "" {
		tsStart, _ := time.Parse(time.RFC3339, CanaryTime)
		tsEnd, err := time.Parse(time.RFC3339, EndTime)
		if err != nil {
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

	configIdLookupURL := fmt.Sprintf(configIdLookupURLFormat, metric.Provider.OPSMX.GateUrl)

	if err := basicChecks(metric); err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	canaryStartTime, baselineStartTime, lifetimeHours, err := getTimeVariables(metric.Provider.OPSMX.BaselineStartTime, metric.Provider.OPSMX.CanaryStartTime, metric.Provider.OPSMX.EndTime, metric.Provider.OPSMX.LifetimeHours)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	var (
		payload        jobPayload
		baselinelog    string
		canarylog      string
		baselinemetric string
		canarymetric   string
	)
	payload.Application = metric.Provider.OPSMX.Application
	payload.CanaryConfig.CanaryHealthCheckHandler.MinimumCanaryResultScore = fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Marginal)
	payload.CanaryConfig.CanarySuccessCriteria.CanaryResultScore = fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Pass)
	payload.CanaryConfig.LifetimeHours = lifetimeHours
	payload.CanaryDeployments = make([]canaryDeployments, 0)

	if metric.Provider.OPSMX.Services != nil {
		baselinelog, canarylog, baselinemetric, canarymetric, err = getServicePayload(metric)
		if err != nil {
			return metricutil.MarkMeasurementError(newMeasurement, err)
		}
		fmt.Printf("%s", baselinemetric)
		if baselinelog != "" && baselinemetric != "" {

			internal := canaryDeployments{
				BaselineStartTimeMs: baselineStartTime,
				CanaryStartTimeMs:   canaryStartTime,
				Baseline:            &logMetric{Log: baselinelog, Metric: baselinemetric},
				Canary:              &logMetric{Log: canarylog, Metric: canarymetric},
			}
			payload.CanaryDeployments = append(payload.CanaryDeployments, internal)
		}
		if baselinelog != "" && baselinemetric == "" {
			internal := canaryDeployments{
				BaselineStartTimeMs: baselineStartTime,
				CanaryStartTimeMs:   canaryStartTime,
				Baseline:            &logMetric{Log: baselinelog},
				Canary:              &logMetric{Log: canarylog},
			}
			payload.CanaryDeployments = append(payload.CanaryDeployments, internal)
		}
		if baselinelog == "" && baselinemetric != "" {
			internal := canaryDeployments{
				BaselineStartTimeMs: baselineStartTime,
				CanaryStartTimeMs:   canaryStartTime,
				Baseline:            &logMetric{Metric: baselinemetric},
				Canary:              &logMetric{Metric: canarymetric},
			}
			payload.CanaryDeployments = append(payload.CanaryDeployments, internal)

		}

	} else {

		internal := canaryDeployments{
			BaselineStartTimeMs: baselineStartTime,
			CanaryStartTimeMs:   canaryStartTime,
		}
		payload.CanaryDeployments = append(payload.CanaryDeployments, internal)
	}
	fmt.Printf("\n\n%v\n\n", payload)
	buffer, err := json.Marshal(payload)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	fmt.Printf("\n%s\n", string(buffer))
	// create a request object
	data, err := makeRequest("POST", configIdLookupURL, string(buffer), metric.Provider.OPSMX.User)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	if !json.Valid(data) {
		err = errors.New("invalid Response")
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	var canaryId string
	var canary map[string]interface{}

	json.Unmarshal(data, &canary)
	if canary["message"] == nil {
		canaryId = fmt.Sprintf("%#v", canary["canaryId"])
	} else {
		str1 := fmt.Sprintf("%#v", canary["message"])
		if len(strings.Split(str1, "message")) > 1 {
			str1 = strings.Split(strings.Split(str1, "message")[1], ",")[0]
			re, _ := regexp.Compile(`[^\w]`)
			str1 = re.ReplaceAllString(str1, " ")
			str1 = strings.TrimSpace(str1)
			str1 = strings.ReplaceAll(str1, "   ", " ")
			err = errors.New(str1)
			return metricutil.MarkMeasurementError(newMeasurement, err)
		} else {
			err = errors.New(str1)
			return metricutil.MarkMeasurementError(newMeasurement, err)
		}
	}
	reportUrl := fmt.Sprintf(reportUrlFormat, metric.Provider.OPSMX.GateUrl, metric.Provider.OPSMX.Application, canaryId)
	fmt.Printf("\n\n%s\n\n", reportUrl)
	//creating a map to return the reporturl and associated data
	mapMetadata := make(map[string]string)
	mapMetadata["canaryId"] = fmt.Sprintf("%v", canaryId)
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
	scoreURL := fmt.Sprintf(scoreUrlFormat, metric.Provider.OPSMX.GateUrl, canaryId)

	data, err := makeRequest("GET", scoreURL, "", metric.Provider.OPSMX.User)
	if err != nil {
		return metricutil.MarkMeasurementError(measurement, err)
	}
	var status map[string]interface{}
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
