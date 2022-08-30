package opsmx

import "time"

const (
	ProviderType                          = "opsmx"
	configIdLookupURLFormat               = `%s/autopilot/api/v3/registerCanary`
	scoreUrlFormat                        = `%s/autopilot/canaries/%s`
	reportUrlFormat                       = `%sui/application/deploymentverification/%s/%s`
	httpConnectionTimeout   time.Duration = 15 * time.Second
	DefaultjobPayloadFormat               = `{
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
	JobPayloadwServices = `{
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

	ServicesjobPayloadFormat = `"canary":{
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
