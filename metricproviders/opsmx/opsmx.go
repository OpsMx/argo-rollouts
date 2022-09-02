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

type Provider struct {
	logCtx log.Entry
	client http.Client
}

// Type indicates provider is a OPSMX provider
func (p *Provider) Type() string {
	return ProviderType
}

// GetMetadata returns any additional metadata which needs to be stored & displayed as part of the metrics result.
func (p *Provider) GetMetadata(metric v1alpha1.Metric) map[string]string {
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
		return []byte{0}, err
	}

	// add a request header
	req.Header.Add("x-spinnaker-user", user)
	req.Header.Add("Content-Type", "application/json")

	// send an HTTP using `req` object
	res, err := http.DefaultClient.Do(req)
	// check for response error
	if err != nil {
		return []byte{0}, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return []byte{0}, err
	}
	return data, err
}

func getServicePayload(metric v1alpha1.Metric) (string, error) {

	var (
		baselinelog    string
		canarylog      string
		baselinemetric string
		canarymetric   string

		ServiceJobPayload     string
		baselinemetricPayload string
		canarymetricPayload   string
		baselinelogPayload    string
		canarylogPayload      string
		baselinePayload       string
		canaryPayload         string

		err error
	)

	for _, item := range metric.Provider.OPSMX.Services {
		if item.LogScopeVariables != "" {
			if len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.BaselineLogScope, ",")) || len(strings.Split(item.LogScopeVariables, ",")) != len(strings.Split(item.CanaryLogScope, ",")) {
				err = errors.New("mismatch in amount of log scope variables and baseline/canary log scope")
				return "", err
			}
			if baselinelog == "" {
				baselinelog = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.BaselineLogScope)
				canarylog = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.CanaryLogScope)
			} else {
				temp := fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.BaselineLogScope)
				baselinelog = fmt.Sprintf("%s\n,\n%s", baselinelog, temp)
				temp = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.LogScopeVariables, item.CanaryLogScope)
				canarylog = fmt.Sprintf("%s\n,\n%s", canarylog, temp)
			}
		}
		if item.MetricScopeVariables != "" {
			if len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.BaselineMetricScope, ",")) || len(strings.Split(item.MetricScopeVariables, ",")) != len(strings.Split(item.CanaryMetricScope, ",")) {
				err = errors.New("mismatch in amount of metric scope variables and baseline/canary metric scope")
				return "", err
			}
			if baselinemetric == "" {
				baselinemetric = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.BaselineMetricScope)
				canarymetric = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.CanaryMetricScope)
			} else {
				temp := fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.BaselineMetricScope)
				baselinemetric = fmt.Sprintf("%s\n,\n%s", baselinemetric, temp)
				temp = fmt.Sprintf(internalFormat, item.ServiceName, item.GateName, item.MetricScopeVariables, item.CanaryMetricScope)
				canarymetric = fmt.Sprintf("%s\n,\n%s", canarymetric, temp)
			}
		}
	}

	if baselinelog != "" {
		baselinelogPayload = fmt.Sprintf(logPayloadFormat, baselinelog)
		canarylogPayload = fmt.Sprintf(logPayloadFormat, canarylog)
	}

	if baselinemetric != "" {
		baselinemetricPayload = fmt.Sprintf(metricPayloadFormat, baselinemetric)
		canarymetricPayload = fmt.Sprintf(metricPayloadFormat, canarymetric)
	}

	if baselinelogPayload == "" && baselinemetricPayload == "" {
		err = errors.New("either provide log or metric arguments")
		return "", err
	}

	if baselinelogPayload != "" && baselinemetricPayload != "" {
		baselinePayload = fmt.Sprintf("%s,\n%s", baselinelogPayload, baselinemetricPayload)
		canaryPayload = fmt.Sprintf("%s,\n%s", canarylogPayload, canarymetricPayload)
		ServiceJobPayload = fmt.Sprintf(servicesjobPayloadFormat, canaryPayload, baselinePayload)
	} else {
		if baselinelogPayload != "" && baselinemetricPayload == "" {
			ServiceJobPayload = fmt.Sprintf(servicesjobPayloadFormat, canarylogPayload, baselinelogPayload)
		} else {
			ServiceJobPayload = fmt.Sprintf(servicesjobPayloadFormat, canarymetricPayload, baselinemetricPayload)
		}
	}
	return ServiceJobPayload, err
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
	var err error

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

	return canaryStartTime, baselineStartTime, lifetimeHours, err
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

	var jobPayload string
	if metric.Provider.OPSMX.Services == nil {
		jobPayload = fmt.Sprintf(defaultjobPayloadFormat, metric.Provider.OPSMX.Application, lifetimeHours, fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Marginal), fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Pass), canaryStartTime, baselineStartTime) //Make the payload
	} else {
		ServiceJobPayload, err := getServicePayload(metric)
		if err != nil {
			return metricutil.MarkMeasurementError(newMeasurement, err)
		}
		jobPayload = fmt.Sprintf(jobPayloadwServices, metric.Provider.OPSMX.Application, lifetimeHours, fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Marginal), fmt.Sprintf("%d", metric.Provider.OPSMX.Threshold.Pass), canaryStartTime, baselineStartTime, ServiceJobPayload)
	}
	// create a request object
	data, err := makeRequest("POST", configIdLookupURL, jobPayload, metric.Provider.OPSMX.User)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	var canaryId string
	var canary map[string]interface{}
	checkvalid := json.Valid(data)
	if checkvalid {
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
	} else {
		err = errors.New("invalid Response")
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}
	reportUrl := fmt.Sprintf(reportUrlFormat, metric.Provider.OPSMX.GateUrl, metric.Provider.OPSMX.Application, canaryId)
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

func (p *Provider) Resume(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	var (
		canaryScore string
		result      map[string]interface{}
		finalScore  map[string]interface{}
	)
	canaryId := measurement.Metadata["canaryId"]
	scoreURL := fmt.Sprintf(scoreUrlFormat, metric.Provider.OPSMX.GateUrl, canaryId)

	data, err := makeRequest("GET", scoreURL, "", metric.Provider.OPSMX.User)
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
	//res.Body.Close()
	if !json.Valid(data) {
		err = errors.New("invalid Response")
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
