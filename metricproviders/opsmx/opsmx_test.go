package opsmx

import (
	"net/http"
	"testing"
	"time"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tj/assert"
)

const (
	httpConnectionTestTimeout time.Duration = 15 * time.Second
)

func TestRunSuccessfully(t *testing.T) {
	// Test Cases
	var tests = []struct {
		metric        v1alpha1.Metric
		expectedPhase v1alpha1.AnalysisPhase
		expectedValue string
	}{
		//Test case for basic function
		{
			metric: v1alpha1.Metric{
				Name: "testappy",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "user1",
						Application:       "testapp",
						BaselineStartTime: "2022-07-29T13:15:00Z",
						CanaryStartTime:   "2022-07-29T13:15:00Z",
						LifetimeHours:     "0.5",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for endTime feature
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "testapp",
						BaselineStartTime: "2022-08-10T13:15:00Z",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for only 1 time format given
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "testapp",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for Single Service feature
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service1",
								GateName:             "gate1",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-datascience-br",
								CanaryMetricScope:    "oes-datascience-cr",
							},
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for multi-service feature
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						User:              "admin",
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service1",
								GateName:             "gate1",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-sapor-br",
								CanaryMetricScope:    "oes-sapor-cr",
							},
							{
								ServiceName:          "service2",
								GateName:             "gate2",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-platform-br",
								CanaryMetricScope:    "oes-platform-cr",
							},
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for multi-service feature along with logs+metrics analysis
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service1",
								GateName:             "gate1",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-platform-br",
								CanaryMetricScope:    "oes-platform-cr",
							},
							{
								ServiceName:          "service2",
								GateName:             "gate2",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-sapor-br",
								CanaryMetricScope:    "oes-sapor-cr",
								LogScopeVariables:    "kubernetes.container_name",
								BaselineLogScope:     "oes-datascience-br",
								CanaryLogScope:       "oes-datascience-cr",
							},
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
		//Test case for 1 incorrect service and one correct
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service1",
								GateName:             "gate1",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-platform-br",
								CanaryMetricScope:    "oes-platform-cr",
							},
							{
								ServiceName:          "service3",
								GateName:             "gate2",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-sapor-br",
								CanaryMetricScope:    "oes-sapor-cr",
								LogScopeVariables:    "kubernetes.container_name",
								BaselineLogScope:     "oes-datascience-br",
								CanaryLogScope:       "oes-datascience-cr",
							},
						},
					},
				},
			},
			expectedValue: "100",
			expectedPhase: v1alpha1.AnalysisPhaseSuccessful,
		},
	}
	for _, test := range tests {
		e := log.NewEntry(log.New())
		c := NewTestHttpClient()
		provider := NewOPSMXProvider(*e, c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		time.Sleep(resumeAfter)
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		// Phase specific cases
		switch test.expectedPhase {
		case v1alpha1.AnalysisPhaseSuccessful:
			assert.NotNil(t, measurement.StartedAt)
			assert.Equal(t, test.expectedValue, measurement.Value)
			assert.Equal(t, test.expectedPhase, measurement.Phase)
			assert.NotNil(t, measurement.FinishedAt)
		case v1alpha1.AnalysisPhaseFailed:
			assert.NotNil(t, measurement.StartedAt)
			assert.Equal(t, test.expectedPhase, measurement.Phase)
			assert.NotNil(t, measurement.FinishedAt)
		}
	}
	var testing = []struct {
		metric        v1alpha1.Metric
		expectedPhase v1alpha1.AnalysisPhase
	}{
		//Test case for Now feature
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:       "https://ds312.isd-dev.opsmx.net/",
						User:          "admin",
						Application:   "testapp",
						LifetimeHours: "0.05",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
					},
				},
			},
		},
		//Test Case for No logs configured still passed in service
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "multiservice",
						User:              "admin",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service1",
								GateName:             "gate1",
								LogScopeVariables:    "kubernetes.container_name",
								BaselineLogScope:     "oes-datascience-br",
								CanaryLogScope:       "oes-datascience-cr",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-sapor-br",
								CanaryMetricScope:    "oes-sapor-cr",
							},
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for incorrect service name
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName:          "service3",
								GateName:             "gate1",
								MetricScopeVariables: "job_name",
								BaselineMetricScope:  "oes-datascience-br",
								CanaryMetricScope:    "oes-datascience-cr",
							},
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseFailed,
		},
		//Test case for no logs and metric analysis given while service was given
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						User:              "admin",
						Application:       "multiservice",
						BaselineStartTime: "",
						CanaryStartTime:   "2022-08-10T13:15:00Z",
						EndTime:           "2022-08-10T13:45:10Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
						Services: []v1alpha1.OPSMXService{
							{
								ServiceName: "service1",
								GateName:    "gate1",
							},
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for no lifetimeHours, Baseline/Canary start time
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:     "https://ds312.isd-dev.opsmx.net/",
						Application: "testapp",
						User:        "admin",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 65,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for Pass score less than marginal
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testapp",
						User:              "admin",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						LifetimeHours:     "0.05",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     60,
							Marginal: 80,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for inappropriate time format
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testapp",
						User:              "admin",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-O8-02T13:15:00Z",
						LifetimeHours:     "0.05",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for incorrect application name
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testap",
						User:              "admin",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						LifetimeHours:     "0.05",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for no lifetimeHours & EndTime
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testapp",
						User:              "admin",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for incorrect gate URL
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmxx.net/",
						Application:       "testapp",
						User:              "admin",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for no user given
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testapp",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
		//Test case for endTime earlier than start time
		{
			metric: v1alpha1.Metric{
				Name: "testapp",
				Provider: v1alpha1.MetricProvider{
					OPSMX: &v1alpha1.OPSMXMetric{
						GateUrl:           "https://ds312.isd-dev.opsmx.net/",
						Application:       "testapp",
						BaselineStartTime: "2022-08-02T13:15:00Z",
						CanaryStartTime:   "2022-08-02T13:15:00Z",
						EndTime:           "2022-08-02T12:45:00Z",
						Threshold: v1alpha1.OPSMXThreshold{
							Pass:     80,
							Marginal: 60,
						},
					},
				},
			},
			expectedPhase: v1alpha1.AnalysisPhaseError,
		},
	}
	for _, test := range testing {
		e := log.NewEntry(log.New())
		c := NewTestHttpClient()
		provider := NewOPSMXProvider(*e, c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		time.Sleep(resumeAfter)
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		// Phase specific cases
		switch test.expectedPhase {
		case v1alpha1.AnalysisPhaseSuccessful:
			assert.NotNil(t, measurement.StartedAt)
			assert.NotNil(t, measurement.FinishedAt)
		case v1alpha1.AnalysisPhaseFailed:
			assert.NotNil(t, measurement.StartedAt)
			assert.Equal(t, test.expectedPhase, measurement.Phase)
			assert.NotNil(t, measurement.FinishedAt)
		case v1alpha1.AnalysisPhaseError:
			assert.NotNil(t, measurement.StartedAt)
			assert.Equal(t, test.expectedPhase, measurement.Phase)
			assert.NotNil(t, measurement.FinishedAt)
		}
	}

}

func newAnalysisRun() *v1alpha1.AnalysisRun {
	return &v1alpha1.AnalysisRun{}
}

func NewTestHttpClient() http.Client {
	c := http.Client{
		Timeout: httpConnectionTestTimeout,
	}
	return c
}