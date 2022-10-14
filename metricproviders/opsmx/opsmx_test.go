package opsmx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/tj/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

var successfulTests = []struct {
	metric                v1alpha1.Metric
	payloadRegisterCanary string
	reportUrl             string
}{
	//Test case for basic function of Single Service feature
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					LifetimeMinutes:   30,
					IntervalTime:      3,
					Delay:             1,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-datascience-br",
							CanaryMetricScope:     "oes-datascience-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
					"lifetimeMinutes": "30",
					"lookBackType": "growing",
					"interval": "3",
					"delay": "1",
					"canaryHealthCheckHandler": {
									"minimumCanaryResultScore": "65"
									},
					"canarySuccessCriteria": {
								"canaryResultScore": "80"
									}
					},
			"canaryDeployments": [
						{
						"canaryStartTimeMs": "1660137300000",
						"baselineStartTimeMs": "1660137300000",
						"canary": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-cr","template":"metricTemplate","templateVersion":"1"}
						  }},
						"baseline": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-br","template":"metricTemplate","templateVersion":"1"}}
						  }
						}
			  ]
		}`,
		reportUrl: "",
	},
	//Test case for endtime function of Single Service feature
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-datascience-br",
							CanaryMetricScope:     "oes-datascience-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
					"lifetimeMinutes": "30",
					"canaryHealthCheckHandler": {
									"minimumCanaryResultScore": "65"
									},
					"canarySuccessCriteria": {
								"canaryResultScore": "80"
									}
					},
			"canaryDeployments": [
						{
						"canaryStartTimeMs": "1660137300000",
						"baselineStartTimeMs": "1660137300000",
						"canary": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-cr","template":"metricTemplate","templateVersion":"1"}
						  }},
						"baseline": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-br","template":"metricTemplate","templateVersion":"1"}}
						  }
						}
			  ]
		}`,
		reportUrl: "",
	},
	//Test case for only 1 time stamp given function of Single Service feature
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-datascience-br",
							CanaryMetricScope:     "oes-datascience-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
					"lifetimeMinutes": "30",
					"canaryHealthCheckHandler": {
									"minimumCanaryResultScore": "65"
									},
					"canarySuccessCriteria": {
								"canaryResultScore": "80"
									}
					},
			"canaryDeployments": [
						{
						"canaryStartTimeMs": "1660137300000",
						"baselineStartTimeMs": "1660137300000",
						"canary": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-cr","template":"metricTemplate","templateVersion":"1"}
						  }},
						"baseline": {
							"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-br","template":"metricTemplate","templateVersion":"1"}}
						  }
						}
			  ]
		}`,
		reportUrl: "",
	},
	//Test case for multi-service feature
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					User:                 "admin",
					GateUrl:              "https://opsmx.test.tst",
					Application:          "multiservice",
					BaselineStartTime:    "2022-08-10T13:15:00Z",
					CanaryStartTime:      "2022-08-10T13:15:00Z",
					EndTime:              "2022-08-10T13:45:10Z",
					GlobalMetricTemplate: "metricTemplate",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-sapor-br",
							CanaryMetricScope:     "oes-sapor-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `		{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
				"lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate"
					}
				  }
				},
				"baseline": {
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},

	//Test case for multi-service feature along with logs+metrics analysis
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-platform-br",
							CanaryMetricScope:     "oes-platform-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
						},
						{
							MetricScopeVariables:  "job_name",
							BaselineMetricScope:   "oes-sapor-br",
							CanaryMetricScope:     "oes-sapor-cr",
							MetricTemplateName:    "metricTemplate",
							MetricTemplateVersion: "1",
							LogScopeVariables:     "kubernetes.container_name",
							BaselineLogScope:      "oes-datascience-br",
							CanaryLogScope:        "oes-datascience-cr",
							LogTemplateName:       "logTemplate",
							LogTemplateVersion:    "1",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
			  "lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-cr",
					  "template":"logTemplate",
					  "templateVersion":"1"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					}
				  }
				},
				"baseline": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-br",
					  "template":"logTemplate",
					  "templateVersion":"1"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate",
					  "templateVersion":"1"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},
	//Test case for 1 incorrect service and one correct
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metricTemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metricTemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logTemplate",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
				"lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-cr",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate"
					}
				  }
				},
				"baseline": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-br",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},
	//Test case for Service Name given
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							ServiceName:          "service1",
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metricTemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metricTemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logTemplate",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
				"lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-cr",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate"
					}
				  }
				},
				"baseline": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-br",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},
	//Test case for Global log Template
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					GlobalLogTemplate: "logTemplate",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							ServiceName:          "service1",
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metricTemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metricTemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",
			"canaryConfig": {
				"lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-cr",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate"
					}
				  }
				},
				"baseline": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-br",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},
	//Test case for CanaryStartTime not given but baseline was given
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							ServiceName:          "service1",
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metricTemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metricTemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logTemplate",
						},
					},
				},
			},
		},
		payloadRegisterCanary: `{
			"application": "multiservice",
			"sourceName":"sourcename",
			"sourceType":"argocd",	
			"canaryConfig": {
				"lifetimeMinutes": "30",
			  "canaryHealthCheckHandler": {
				"minimumCanaryResultScore": "65"
			  },
			  "canarySuccessCriteria": {
				"canaryResultScore": "80"
			  }
			},
			"canaryDeployments": [
			  {
				"canaryStartTimeMs": "1660137300000",
				"baselineStartTimeMs": "1660137300000",
				"canary": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-cr",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-cr",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-cr",
					  "template":"metricTemplate"
					}
				  }
				},
				"baseline": {
				  "log": {
					"service2": {
					  "serviceGate": "gate2",
					  "kubernetes.container_name": "oes-datascience-br",
					  "template":"logTemplate"
					}
				  },
				  "metric": {
					"service1": {
					  "serviceGate": "gate1",
					  "job_name": "oes-platform-br",
					  "template":"metricTemplate"
					},
					"service2": {
					  "serviceGate": "gate2",
					  "job_name": "oes-sapor-br",
					  "template":"metricTemplate"
					}
				  }
				}
			  }
			]
		  }`,
		reportUrl: "",
	},
}

var negativeTests = []struct {
	metric        v1alpha1.Metric
	expectedPhase v1alpha1.AnalysisPhase
	message       string
}{

	//Test case for no lifetimeMinutes, Baseline/Canary start time
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:     "https://opsmx.test.tst",
					Application: "testapp",
					User:        "admin",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "either provide lifetimeMinutes or end time",
	},
	//Test case for Pass score less than marginal
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					User:              "admin",
					BaselineStartTime: "2022-08-02T13:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   30,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     60,
						Marginal: 80,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "pass score cannot be less than marginal score",
	},
	//Test case for inappropriate time format canary
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					User:              "admin",
					BaselineStartTime: "2022-08-02T13:15:00Z",
					CanaryStartTime:   "2022-O8-02T13:15:00Z",
					LifetimeMinutes:   30,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "parsing time \"2022-O8-02T13:15:00Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"O8-02T13:15:00Z\" as \"01\"",
	},
	//Test case for inappropriate time format baseline
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					User:              "admin",
					BaselineStartTime: "2022-O8-02T13:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   30,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "parsing time \"2022-O8-02T13:15:00Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"O8-02T13:15:00Z\" as \"01\"",
	},
	//Test case for inappropriate time format endTime
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					User:              "admin",
					BaselineStartTime: "2022-08-02T13:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					EndTime:           "2022-O8-02T13:15:00Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "parsing time \"2022-O8-02T13:15:00Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"O8-02T13:15:00Z\" as \"01\"",
	},
	//Test case for no lifetimeMinutes & EndTime
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					User:              "admin",
					BaselineStartTime: "2022-08-02T13:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},
		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "either provide lifetimeMinutes or end time",
	},
	//Test case for No log & Metric analysis
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
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

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "no services provided",
	},
	//Test case for No log & Metric analysis
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							ServiceName: "service1",
						},
						{
							ServiceName: "service2",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "at least one of log or metric context must be included",
	},
	//Test case for mismatch in log scope variables and baseline/canary log scope
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metrictemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metrictemplate",
							LogScopeVariables:    "kubernetes.container_name,kubernetes.pod",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "mismatch in number of log scope variables and baseline/canary log scope",
	},

	//Test case for mismatch in metric scope variables and baseline/canary metric scope
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					User:              "admin",
					Application:       "multiservice",
					BaselineStartTime: "2022-08-10T13:15:00Z",
					CanaryStartTime:   "2022-08-10T13:15:00Z",
					EndTime:           "2022-08-10T13:45:10Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name,job123",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metrictemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metrictemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "mismatch in number of metric scope variables and baseline/canary metric scope",
	},
	//Test case for when end time is less than start time
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T13:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					EndTime:           "2022-08-02T12:45:00Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "start time cannot be greater than end time",
	},
	//Test case when end time given and baseline and canary start time not same
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					EndTime:           "2022-08-02T12:45:00Z",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "both start time should be kept same in case of using end time argument",
	},
	//Test case when lifetimeMinutes is less than 3 minutes
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   2,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "lifetime minutes cannot be less than 3 minutes",
	},
	//Test case when intervalTime is less than 3 minutes
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      2,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-datascience-br",
							CanaryMetricScope:    "oes-datascience-cr",
							MetricTemplateName:   "metrictemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "interval time cannot be less than 3 minutes",
	},
	//Test case when baseline or canary logplaceholder is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metrictemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-cr",
							LogTemplateName:      "logtemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing canary for log analysis",
	},
	//Test case when baseline or canary metricplaceholder is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metrictemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemplate",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing baseline/canary for metric analysis",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metrictemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "provide either a service specific log template or global log template",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "provide either a service specific metric template or global metric template",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing log Scope placeholder for the provided baseline/canary",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing log Scope placeholder for the provided baseline/canary",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							BaselineLogScope:     "oes-datascienece-br",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing log Scope placeholder for the provided baseline/canary",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							BaselineMetricScope: "oes-sapor-br",
							CanaryMetricScope:   "oes-sapor-cr",
							LogScopeVariables:   "kubernetes.container_name",
							BaselineLogScope:    "oes-datascienece-br",
							CanaryLogScope:      "oes-datascience-cr",
							LogTemplateName:     "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing metric Scope placeholder for the provided baseline/canary",
	},
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							CanaryMetricScope: "oes-sapor-cr",
							LogScopeVariables: "kubernetes.container_name",
							BaselineLogScope:  "oes-datascienece-br",
							CanaryLogScope:    "oes-datascience-cr",
							LogTemplateName:   "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing metric Scope placeholder for the provided baseline/canary",
	},
	//Test case when global and service specific template is missing
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							BaselineMetricScope: "oes-sapor-br",
							LogScopeVariables:   "kubernetes.container_name",
							BaselineLogScope:    "oes-datascienece-br",
							CanaryLogScope:      "oes-datascience-cr",
							LogTemplateName:     "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "missing metric Scope placeholder for the provided baseline/canary",
	},
	//Test case when intervalTime is given but lookBackType is not given
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "prom",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "interval time is given and lookbacktype is required to run interval analysis",
	},
	//Test case when intervalTime is not given but lookBackType is given
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "https://opsmx.test.tst",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "prom",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "lookbacktype is given and interval time is required to run interval analysis",
	},
	//Test case when improper URL
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:           "	",
					Application:       "testapp",
					BaselineStartTime: "2022-08-02T14:15:00Z",
					CanaryStartTime:   "2022-08-02T13:15:00Z",
					LifetimeMinutes:   60,
					IntervalTime:      3,
					LookBackType:      "growing",
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 60,
					},
					Services: []v1alpha1.OPSMXService{
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "prom",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascienece-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logtemp",
						},
					},
				},
			},
		},

		expectedPhase: v1alpha1.AnalysisPhaseError,
		message:       "parse \"\\t\": net/url: invalid control character in URL",
	},
}

var nowFeature = []struct {
	metric    v1alpha1.Metric
	reportUrl string
}{
	//Test case for CanaryStartTime not given but baseline was given
	{
		metric: v1alpha1.Metric{
			Name: "testapp",
			Provider: v1alpha1.MetricProvider{
				OPSMX: &v1alpha1.OPSMXMetric{
					GateUrl:         "https://opsmx.test.tst",
					User:            "admin",
					Application:     "multiservice",
					LifetimeMinutes: 6,
					Threshold: v1alpha1.OPSMXThreshold{
						Pass:     80,
						Marginal: 65,
					},
					Services: []v1alpha1.OPSMXService{
						{
							ServiceName:          "service1",
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-platform-br",
							CanaryMetricScope:    "oes-platform-cr",
							MetricTemplateName:   "metricTemplate",
						},
						{
							MetricScopeVariables: "job_name",
							BaselineMetricScope:  "oes-sapor-br",
							CanaryMetricScope:    "oes-sapor-cr",
							MetricTemplateName:   "metricTemplate",
							LogScopeVariables:    "kubernetes.container_name",
							BaselineLogScope:     "oes-datascience-br",
							CanaryLogScope:       "oes-datascience-cr",
							LogTemplateName:      "logTemplate",
						},
					},
				},
			},
		},
		reportUrl: "",
	},
}

const (
	endpointRegisterCanary    = "https://opsmx.test.tst/autopilot/api/v5/registerCanary"
	endpointCheckCanaryStatus = "https://opsmx.test.tst/autopilot/canaries/1424"
)

func getFakeClient(dataParam map[string][]byte) *k8sfake.Clientset {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"gate-url":       []byte("https://opsmx.secret.tst"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("admin"),
	}
	if len(dataParam) != 0 {
		data = dataParam
	}
	opsmxSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultSecretName,
		},
		Data: data,
	}
	fakeClient := k8sfake.NewSimpleClientset()
	fakeClient.PrependReactor("get", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, opsmxSecret, nil
	})

	return fakeClient
}

func TestRunSucessCases(t *testing.T) {
	// Test Cases
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, endpointRegisterCanary, req.URL.String())

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			bodyI := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyI)
			if err != nil {
				panic(err)
			}
			expectedBodyI := map[string]interface{}{}
			err = json.Unmarshal([]byte(test.payloadRegisterCanary), &expectedBodyI)
			if err != nil {
				panic(err)
			}
			assert.Equal(t, expectedBodyI, bodyI)
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"canaryId": 1424
				}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})
		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		assert.NotNil(t, measurement.StartedAt)
		assert.Equal(t, "1424", measurement.Metadata["canaryId"])
		assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
	}
}

func TestResumeSucessCases(t *testing.T) {

	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 100,
						"intervalNo": 1,
						"isLastRun": true,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "COMPLETED"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"

		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, "100", measurement.Value)
		assert.NotNil(t, measurement.FinishedAt)
		assert.Equal(t, "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424", measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseSuccessful, measurement.Phase)
	}
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 0,
						"intervalNo": 2,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "COMPLETED"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"

		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, "0", measurement.Value)
		assert.NotNil(t, measurement.FinishedAt)
		assert.Equal(t, "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424", measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseFailed, measurement.Phase)
		if test.metric.Provider.OPSMX.LookBackType != "" {
			assert.Equal(t, "Interval Analysis Failed at intervalNo. 2", measurement.Metadata["interval analysis message"])
			assert.Equal(t, "2", measurement.Metadata["Current intervalNo"])
		}
	}
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 75,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "COMPLETED"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"

		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, "75", measurement.Value)
		assert.NotNil(t, measurement.FinishedAt)
		assert.Equal(t, "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424", measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseInconclusive, measurement.Phase)
	}
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 75,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "CANCELLED"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"

		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.NotNil(t, measurement.FinishedAt)
		assert.Equal(t, "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424", measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseFailed, measurement.Phase)
	}
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 75,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "RUNNING"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"

		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
		assert.Equal(t, "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424", measurement.Metadata["reportUrl"])
	}
}

func TestFailNoLogsConfiguredStillPassedInService(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())

		return &http.Response{
			StatusCode: 404,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			{
				"timestamp": 1.662356583464E12,
				"status": 404.0,
				"error": "Not Found",
				"message": "Log template not configured for a service : service1",
				"path": "/autopilot/api/v3/registerCanary"
			}
			`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "multiservice",
				User:              "admin",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				EndTime:           "2022-08-10T13:45:10Z",
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						LogScopeVariables:    "kubernetes.container_name",
						BaselineLogScope:     "oes-datascience-br",
						CanaryLogScope:       "oes-datascience-cr",
						LogTemplateName:      "logtemplate",
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-sapor-br",
						CanaryMetricScope:    "oes-sapor-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}

	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "Error: Not Found\nMessage: Log template not configured for a service : service1", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)

}

func TestNowFeature(t *testing.T) {
	for _, test := range nowFeature {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, endpointRegisterCanary, req.URL.String())

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			bodyI := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyI)
			if err != nil {
				panic(err)
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"canaryId": 1424
				}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})
		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		assert.NotNil(t, measurement.StartedAt)
		assert.Equal(t, "1424", measurement.Metadata["canaryId"])
		assert.Equal(t, test.reportUrl, measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
	}
}

func TestRolloutFunctions(t *testing.T) {
	for _, test := range nowFeature {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, endpointRegisterCanary, req.URL.String())

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			bodyI := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyI)
			if err != nil {
				panic(err)
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"canaryId": 1424
				}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})
		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		assert.NotNil(t, measurement.StartedAt)
		assert.Equal(t, "1424", measurement.Metadata["canaryId"])
		assert.Equal(t, fmt.Sprintf("%s", test.reportUrl), measurement.Metadata["reportUrl"])
		assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
		measurement2 := provider.Terminate(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, measurement, measurement2)
		measurement3 := provider.GarbageCollect(newAnalysisRun(), test.metric, 0)
		assert.Equal(t, nil, measurement3)
		assert.IsType(t, http.Client{}, NewHttpClient())
		assert.Equal(t, "opsmx", provider.Type())
		measurement4 := provider.GetMetadata(test.metric)
		assert.Equal(t, map[string]string(map[string]string(nil)), measurement4)
	}
}

func TestInvalidJson(t *testing.T) {
	for _, test := range successfulTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {

			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
					"owner": "admin",
					"application": "testapp",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-02 10:02:18.504",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/testapp/1424",
						"overallScore": 75,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-02 10:02:18.504",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": false,
						"status": "RUNNING"
					}}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})

		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

		mapMetadata := make(map[string]string)
		mapMetadata["canaryId"] = "1424"
		measurement := v1alpha1.Measurement{
			Metadata: mapMetadata,
			Phase:    v1alpha1.AnalysisPhaseRunning,
		}
		measurement = provider.Resume(newAnalysisRun(), test.metric, measurement)
		assert.Equal(t, "invalid Response", measurement.Message)
		assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
	}
}

func TestIncorrectApplicationName(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())

		return &http.Response{
			StatusCode: 500,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			{
				"timestamp": 1662442034995,
				"status": 500,
				"error": "Internal Server Error",
				"exception": "feign.FeignException$NotFound",
				"message": "Application not found with name testap"
			}
			`)),
			Header: make(http.Header),
		}, nil
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "testap",
				User:              "admin",
				BaselineStartTime: "2022-08-02T13:15:00Z",
				CanaryStartTime:   "2022-08-02T13:15:00Z",
				LifetimeMinutes:   30,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 60,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-datascience-br",
						CanaryMetricScope:    "oes-datascience-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "Error: Internal Server Error\nMessage: Application not found with name testap", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)

}

func TestIncorrectGateURL(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Header:     make(http.Header),
		}, errors.New("Post \"https://opsmx.invalidurl.tst\": dial tcp: lookup https://opsmx.invalidurl.tst: no such host")
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.invalidurl.tst",
				Application:       "testapp",
				User:              "admin",
				BaselineStartTime: "2022-08-02T13:15:00Z",
				CanaryStartTime:   "2022-08-02T13:15:00Z",
				LifetimeMinutes:   30,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 60,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-datascience-br",
						CanaryMetricScope:    "oes-datascience-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)

}

func TestIncorrectScoreURL(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Header:     make(http.Header),
		}, errors.New("Post \"https://opsmx.invalidurl.tst\": dial tcp: lookup https://opsmx.invalidurl.tst: no such host")
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.invalidurl.tst",
				Application:       "testapp",
				User:              "admin",
				BaselineStartTime: "2022-08-02T13:15:00Z",
				CanaryStartTime:   "2022-08-02T13:15:00Z",
				LifetimeMinutes:   30,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 60,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-datascience-br",
						CanaryMetricScope:    "oes-datascience-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
	mapMetadata := make(map[string]string)
	mapMetadata["canaryId"] = "1424"

	measurement := v1alpha1.Measurement{
		Metadata: mapMetadata,
	}
	measurement = provider.Resume(newAnalysisRun(), metric, measurement)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
	fmt.Printf("\n\n")
	fmt.Printf("ScoreUrl:%s", measurement.Metadata["scoreUrl"])
	fmt.Printf("\n\n")
}

func TestNoUserDefined(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			{
				"timestamp": 1662442034995,
				"status": 500,
				"error": "Internal Server Error",
				"exception": "feign.FeignException$NotFound",
				"message": "message1"
			}
			`)),
			Header: make(http.Header),
		}, nil
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "testapp",
				BaselineStartTime: "2022-08-02T13:15:00Z",
				CanaryStartTime:   "2022-08-02T13:15:00Z",
				LifetimeMinutes:   30,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 60,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-datascience-br",
						CanaryMetricScope:    "oes-datascience-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "Error: Internal Server Error\nMessage: message1", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)

}

func TestIncorrectServiceName(t *testing.T) {
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {

		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
					"owner": "admin",
					"application": "multiservice",
					"canaryResult": {
						"duration": "0 seconds",
						"lastUpdated": "2022-09-06 09:16:51.58",
						"canaryReportURL": "https://opsmx.test.tst/ui/application/deploymentverification/multiservice/1424",
						"overallScore": null,
						"overallResult": "HEALTHY",
						"message": "Canary Is HEALTHY",
						"errors": []
					},
					"launchedDate": "2022-09-06 09:16:51.539",
					"canaryConfig": {
						"combinedCanaryResultStrategy": "LOWEST",
						"minimumCanaryResultScore": 65.0,
						"name": "admin",
						"lifetimeMinutes": 30,
						"canaryAnalysisIntervalMins": 30,
						"maximumCanaryResultScore": 80.0
					},
					"id": "1424",
					"services": [],
					"status": {
						"complete": true,
						"status": "COMPLETED"
					}
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst/",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				EndTime:           "2022-08-10T13:45:10Z",
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables: "job_name",
						BaselineMetricScope:  "oes-datascience-br",
						CanaryMetricScope:    "oes-datascience-cr",
						MetricTemplateName:   "metrictemplate",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)

	mapMetadata := make(map[string]string)
	mapMetadata["canaryId"] = "1424"
	mapMetadata["reportUrl"] = "https://opsmx.test.tst/ui/application/deploymentverification/multiservice/1424"

	measurement := v1alpha1.Measurement{
		Metadata: mapMetadata,
		Phase:    v1alpha1.AnalysisPhaseRunning,
	}
	measurement = provider.Resume(newAnalysisRun(), metric, measurement)
	assert.NotEqual(t, "100", measurement.Value)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseFailed, measurement.Phase)
}

func TestGenericNegativeTestsRun(t *testing.T) {
	for _, test := range negativeTests {
		e := log.NewEntry(log.New())
		c := NewTestClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, endpointRegisterCanary, req.URL.String())
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, nil
		})
		provider := NewOPSMXProvider(*e, getFakeClient(map[string][]byte{}), c)
		measurement := provider.Run(newAnalysisRun(), test.metric)
		assert.NotNil(t, measurement.StartedAt)
		assert.NotNil(t, measurement.FinishedAt)
		assert.Equal(t, test.message, measurement.Message)
		assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
	}
}

// Secret is not created
func TestSecretNotCreated(t *testing.T) {
	fakeClient := k8sfake.NewSimpleClientset()
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "secrets \"opsmx-profile\" not found", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Secret is entered in the template
func TestSecretEnteredInTemplate(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"gate-url":       []byte("https://opsmx.secret.tst"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("admin"),
	}
	opsmxSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-secret-name",
		},
		Data: data,
	}
	fakeClient := k8sfake.NewSimpleClientset()
	fakeClient.PrependReactor("get", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, opsmxSecret, nil
	})

	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Profile:           "new-secret-xyz",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
}

// Secret mentioned in the template is not found
func TestSecretTemplateNotFound(t *testing.T) {
	fakeClient := k8sfake.NewSimpleClientset()
	fakeClient.PrependReactor("get", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("secret not found")
	})

	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Profile:           "new-secret-xyz",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "secret not found", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Source Name is not mentioned in the secret
func TestSourceNameMissingInSecret(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "source-name is not specified in the secret", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Gate url not in secret or in the Template
func TestGateUrlMissing(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		// "gate-url": []byte("https://opsmx.secret.tst"),
		"source-name": []byte("sourcename"),
		"user":        []byte("admin"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				// GateUrl:           "https://opsmx.test.tst",
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "the gate-url is not specified both in the template and in the secret", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Gate url is picked from the secret
func TestGateUrlPickedFromSecret(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"gate-url":       []byte("https://opsmx.secret.tst"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("admin"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://opsmx.secret.tst/autopilot/api/v5/registerCanary", req.URL.String())

		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
					{
						"canaryId": 1424
					}
					`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				User:              "admin",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.Nil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
}

// User is picked from the secret
func TestUserPickedFromSecret(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"gate-url":       []byte("https://opsmx.secret.tst"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("usersecret"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://opsmx.secret.tst/autopilot/api/v5/registerCanary", req.URL.String())
		assert.Equal(t, "usersecret", req.Header.Get("x-spinnaker-user"))
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
						{
							"canaryId": 1424
						}
						`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.Nil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
}

// User is picked from the template
func TestUserPickedFromTemplate(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"gate-url":       []byte("https://opsmx.secret.tst"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("usersecret"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://opsmx.secret.tst/autopilot/api/v5/registerCanary", req.URL.String())
		assert.Equal(t, "usertemplate", req.Header.Get("x-spinnaker-user"))
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
							{
								"canaryId": 1424
							}
							`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				User:              "usertemplate",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.Nil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
}

// User is not mentioned both in the secret as well as in the template
func TestUserMissing(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("true"),
		"source-name":    []byte("sourcename"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "the user is not specified both in the template and in the secret", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// The value of CDintegration is neither True nor False
func TestCdIntegrationValueNegative(t *testing.T) {
	data := map[string][]byte{
		"cd-integration": []byte("xyz"),
		"source-name":    []byte("sourcename"),
		"user":           []byte("admin"),
	}
	fakeClient := getFakeClient(data)
	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
					{
					}
					`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})
	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "cd-integration should be either true or false", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Test case for when cdIntegration value is not present
func TestCdIntegrationValueNotPresent(t *testing.T) {
	data := map[string][]byte{
		"source-name": []byte("sourcename"),
		"user":        []byte("admin"),
	}
	fakeClient := getFakeClient(data)

	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`
				{
				}
				`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.NotNil(t, measurement.FinishedAt)
	assert.Equal(t, "cd-integration is not specified in the secret", measurement.Message)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, measurement.Phase)
}

// Test case for when cdIntegration value is set to false
func TestCdIntegrationValueIsFalse(t *testing.T) {
	data := map[string][]byte{
		"source-name":    []byte("sourcename"),
		"user":           []byte("admin"),
		"cd-integration": []byte("false"),
	}
	fakeClient := getFakeClient(data)
	payload := `{
		"application": "multiservice",
		"sourceName":"sourcename",
		"sourceType":"argorollouts",
		"canaryConfig": {
				"lifetimeMinutes": "30",
				"lookBackType": "growing",
				"interval": "3",
				"canaryHealthCheckHandler": {
								"minimumCanaryResultScore": "65"
								},
				"canarySuccessCriteria": {
							"canaryResultScore": "80"
								}
				},
		"canaryDeployments": [
					{
					"canaryStartTimeMs": "1660137300000",
					"baselineStartTimeMs": "1660137300000",
					"canary": {
						"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-cr","template":"metricTemplate","templateVersion":"1"}
					  }},
					"baseline": {
						"metric": {"service1":{"serviceGate":"gate1","job_name":"oes-datascience-br","template":"metricTemplate","templateVersion":"1"}}
					  }
					}
		  ]
	}`

	e := log.NewEntry(log.New())
	c := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, endpointRegisterCanary, req.URL.String())

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		bodyI := map[string]interface{}{}
		err = json.Unmarshal(body, &bodyI)
		if err != nil {
			panic(err)
		}
		expectedBodyI := map[string]interface{}{}
		err = json.Unmarshal([]byte(payload), &expectedBodyI)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, expectedBodyI, bodyI)
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			{
				"canaryId": 1424
			}
			`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})

	metric := v1alpha1.Metric{
		Name: "testapp",
		Provider: v1alpha1.MetricProvider{
			OPSMX: &v1alpha1.OPSMXMetric{
				GateUrl:           "https://opsmx.test.tst",
				Application:       "multiservice",
				BaselineStartTime: "2022-08-10T13:15:00Z",
				CanaryStartTime:   "2022-08-10T13:15:00Z",
				LifetimeMinutes:   30,
				LookBackType:      "growing",
				IntervalTime:      3,
				Threshold: v1alpha1.OPSMXThreshold{
					Pass:     80,
					Marginal: 65,
				},
				Services: []v1alpha1.OPSMXService{
					{
						MetricScopeVariables:  "job_name",
						BaselineMetricScope:   "oes-datascience-br",
						CanaryMetricScope:     "oes-datascience-cr",
						MetricTemplateName:    "metricTemplate",
						MetricTemplateVersion: "1",
					},
				},
			},
		},
	}
	provider := NewOPSMXProvider(*e, fakeClient, c)
	measurement := provider.Run(newAnalysisRun(), metric)
	assert.NotNil(t, measurement.StartedAt)
	assert.Nil(t, measurement.FinishedAt)
	assert.Equal(t, v1alpha1.AnalysisPhaseRunning, measurement.Phase)
}

func newAnalysisRun() *v1alpha1.AnalysisRun {
	return &v1alpha1.AnalysisRun{}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) http.Client {
	return http.Client{
		Transport: fn,
	}
}