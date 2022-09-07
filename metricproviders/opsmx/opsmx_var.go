package opsmx

import (
	"time"
)

const (
	ProviderType                          = "opsmx"
	configIdLookupURLFormat               = `%s/autopilot/api/v3/registerCanary`
	scoreUrlFormat                        = `%s/autopilot/canaries/%s`
	reportUrlFormat                       = `%sui/application/deploymentverification/%s/%s`
	resumeAfter                           = 15 * time.Second
	httpConnectionTimeout   time.Duration = 15 * time.Second
	defaultjobPayloadFormat               = `{
        "application": "%s",
        "canaryConfig": {
                "lifetimeHours": %s,
                "canaryHealthCheckHandler": {
                                "minimumCanaryResultScore": %s
                                },
                "canarySuccessCriteria": {
                            "canaryResultScore": %s
                                }
                },
        "canaryDeployments": [
                    {
                    "canaryStartTimeMs": %s,
                    "baselineStartTimeMs": %s
                    }
          ]
    }`
	jobPayloadwServices = `{
        "application": "%s",
        "canaryConfig": {
                "lifetimeHours": %s,
                "canaryHealthCheckHandler": {
                                "minimumCanaryResultScore": %s
                                },
                "canarySuccessCriteria": {
                            "canaryResultScore": %s
                                }
                },
        "canaryDeployments": [
                    {
                    "canaryStartTimeMs": %s,
                    "baselineStartTimeMs": %s,
					%s
                    }
          ]
    }`

	servicesjobPayloadFormat = `"canary":{
			%s
		},
		"baseline":{
			%s
		}`
	logPayloadFormat = `"log": {
			%s
		}`
	metricPayloadFormat = `"metric": {
			%s
		}`
	internalFormat = `"%s": {
		"serviceGate": "%s",
		"%s": "%s"
		}`
)

type jobPayload struct {
	Application  string       `json:"application,omitempty"`
	CanaryConfig canaryConfig `json:"canaryConfig,omitempty"`
}

type canaryConfig struct {
	LifetimeHours float64 `json:"lifetimeHours,omitempty"`
	Name          string  `json:"name,omitempty"`
}
