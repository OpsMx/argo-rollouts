import { ActionButton, WaitFor } from 'argo-ui/v2';
import * as React from 'react';
import './rollout.scss';
import '../pods/pods.scss';
// import LoadingSpinner from "./LoadingSpinner";


export const ReportsWidget = (props: {  clickback: any; reportsInput: {}}) => {
    const [getURL, setURL] = React.useState('');
    const [loading, setLoading] = React.useState(true);
    const LoadApiCalls = (props: any) => {
        setLoading(true);
        let url2 = '/api/v1/applications/' + props.applicationName + '/resource?name=' + props.resouceName + '&appNamespace=' + props.nameSpace + '&namespace=' + props.nameSpace + '&resourceName=' + props.resouceName + '&version=' + props.version + '&kind=AnalysisRun&group=argoproj.io';
        fetch(url2)
          .then(response => {
            return response.json()
          })
          .then((data: any) => {
            data = {
              "manifest": "{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"AnalysisRun\",\"metadata\":{\"annotations\":{\"rollout.argoproj.io/revision\":\"2\"},\"creationTimestamp\":\"2022-11-10T05:41:00Z\",\"generation\":3,\"labels\":{\"rollout-type\":\"Step\",\"rollouts-pod-template-hash\":\"db465c966\",\"step-index\":\"2\"},\"managedFields\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:metadata\":{\"f:annotations\":{\".\":{},\"f:rollout.argoproj.io/revision\":{}},\"f:labels\":{\".\":{},\"f:rollout-type\":{},\"f:rollouts-pod-template-hash\":{},\"f:step-index\":{}},\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"2a629d52-d9e7-4a40-89cb-3eef3869828b\\\"}\":{}}},\"f:spec\":{\".\":{},\"f:args\":{},\"f:metrics\":{}},\"f:status\":{\".\":{},\"f:dryRunSummary\":{},\"f:message\":{},\"f:metricResults\":{},\"f:phase\":{},\"f:runSummary\":{\".\":{},\"f:count\":{},\"f:failed\":{}},\"f:startedAt\":{}}},\"manager\":\"rollouts-controller\",\"operation\":\"Update\",\"time\":\"2022-11-10T05:44:35Z\"}],\"name\":\"junaid-on-job-db465c966-2-2\",\"namespace\":\"argocd\",\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"blockOwnerDeletion\":true,\"controller\":true,\"kind\":\"Rollout\",\"name\":\"junaid-on-job\",\"uid\":\"2a629d52-d9e7-4a40-89cb-3eef3869828b\"}],\"resourceVersion\":\"51067115\",\"uid\":\"37700755-22b7-482f-b9d9-c26518d06140\"},\"spec\":{\"args\":[{\"name\":\"canary-hash\",\"value\":\"db465c966\"},{\"name\":\"baseline-hash\",\"value\":\"5d4bfb9547\"}],\"metrics\":[{\"name\":\"junaid-job-analysis\",\"provider\":{\"job\":{\"metadata\":{},\"spec\":{\"backoffLimit\":2,\"template\":{\"metadata\":{},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MY_POD_NAME\",\"valueFrom\":{\"fieldRef\":{\"fieldPath\":\"metadata.name\"}}},{\"name\":\"STABLE_POD_HASH\",\"value\":\"{{args.baseline-hash}}\"},{\"name\":\"LATEST_POD_HASH\",\"value\":\"{{args.canary-hash}}\"},{\"name\":\"PROVIDER_CONFIG\",\"value\":\"providerConfig\"}],\"image\":\"genos1998/argojob:v7\",\"name\":\"junaid-job-analysis\",\"resources\":{},\"volumeMounts\":[{\"mountPath\":\"/etc/config/provider\",\"name\":\"provider-config-volume\"},{\"mountPath\":\"/etc/config/secrets\",\"name\":\"secret-config-volume\",\"readOnly\":true}]}],\"restartPolicy\":\"Never\",\"serviceAccountName\":\"job-acc\",\"volumes\":[{\"configMap\":{\"name\":\"opsmx-provider-config\"},\"name\":\"provider-config-volume\"},{\"name\":\"secret-config-volume\",\"secret\":{\"optional\":false,\"secretName\":\"opsmx-profile-v5\"}}]}}}}}}]},\"status\":{\"dryRunSummary\":{},\"message\":\"Metric \\\"junaid-job-analysis\\\" assessed Failed due to failed (1) \\u003e failureLimit (0)\",\"metricResults\":[{\"count\":1,\"failed\":1,\"measurements\":[{\"finishedAt\":\"2022-11-10T05:44:35Z\",\"metadata\":{\"job-name\":\"37700755-22b7-482f-b9d9-c26518d06140.junaid-job-analysis.1\"},\"phase\":\"Failed\",\"startedAt\":\"2022-11-10T05:41:00Z\"}],\"name\":\"junaid-job-analysis\",\"phase\":\"Failed\"}],\"phase\":\"Failed\",\"runSummary\":{\"count\":1,\"failed\":1},\"startedAt\":\"2022-11-10T05:41:00Z\"}}"
            };
            if (data.manifest.includes('job-name')) {
              let b = JSON.parse(data.manifest);
              console.log(b);
              if (b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']) {
                fetchEndpointURL(props.applicationName, props.resouceName, props.nameSpace, props.version, b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']);
              }
            }
          }).catch(err => {
            console.error('res.data', err)
          });
      };

    const fetchEndpointURL = (applicationName: String, resouceName: String, nameSpace: String, version: String, jobName: String) => {
        let url3 = '/api/v1/applications/' + applicationName + '/resource?name=' + jobName + '&appNamespace=' + nameSpace + '&namespace=' + nameSpace + '&resourceName=' + jobName + '&version=v1&kind=Job&group=batch'
        fetch(url3)
          .then(response => {
            return response.json()
          })
          .then((data: any) => {
            data = {
              "manifest": "{\"apiVersion\":\"batch/v1\",\"kind\":\"Job\",\"metadata\":{\"annotations\":{\"analysisrun.argoproj.io/metric-name\":\"junaid-job-analysis\",\"analysisrun.argoproj.io/name\":\"junaid-on-job-db465c966-2-2\"},\"creationTimestamp\":\"2022-11-10T05:41:00Z\",\"generation\":1,\"labels\":{\"analysisrun.argoproj.io/uid\":\"37700755-22b7-482f-b9d9-c26518d06140\"},\"managedFields\":[{\"apiVersion\":\"batch/v1\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:status\":{\"f:active\":{},\"f:startTime\":{}}},\"manager\":\"kube-controller-manager\",\"operation\":\"Update\",\"subresource\":\"status\",\"time\":\"2022-11-10T05:41:00Z\"},{\"apiVersion\":\"batch/v1\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:metadata\":{\"f:annotations\":{\".\":{},\"f:analysisrun.argoproj.io/metric-name\":{},\"f:analysisrun.argoproj.io/name\":{}},\"f:labels\":{\".\":{},\"f:analysisrun.argoproj.io/uid\":{}},\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"37700755-22b7-482f-b9d9-c26518d06140\\\"}\":{}}},\"f:spec\":{\"f:backoffLimit\":{},\"f:completionMode\":{},\"f:completions\":{},\"f:parallelism\":{},\"f:suspend\":{},\"f:template\":{\"f:spec\":{\"f:containers\":{\"k:{\\\"name\\\":\\\"junaid-job-analysis\\\"}\":{\".\":{},\"f:env\":{\".\":{},\"k:{\\\"name\\\":\\\"LATEST_POD_HASH\\\"}\":{\".\":{},\"f:name\":{},\"f:value\":{}},\"k:{\\\"name\\\":\\\"MY_POD_NAME\\\"}\":{\".\":{},\"f:name\":{},\"f:valueFrom\":{\".\":{},\"f:fieldRef\":{}}},\"k:{\\\"name\\\":\\\"PROVIDER_CONFIG\\\"}\":{\".\":{},\"f:name\":{},\"f:value\":{}},\"k:{\\\"name\\\":\\\"STABLE_POD_HASH\\\"}\":{\".\":{},\"f:name\":{},\"f:value\":{}}},\"f:image\":{},\"f:imagePullPolicy\":{},\"f:name\":{},\"f:resources\":{},\"f:terminationMessagePath\":{},\"f:terminationMessagePolicy\":{},\"f:volumeMounts\":{\".\":{},\"k:{\\\"mountPath\\\":\\\"/etc/config/provider\\\"}\":{\".\":{},\"f:mountPath\":{},\"f:name\":{}},\"k:{\\\"mountPath\\\":\\\"/etc/config/secrets\\\"}\":{\".\":{},\"f:mountPath\":{},\"f:name\":{},\"f:readOnly\":{}}}}},\"f:dnsPolicy\":{},\"f:restartPolicy\":{},\"f:schedulerName\":{},\"f:securityContext\":{},\"f:serviceAccount\":{},\"f:serviceAccountName\":{},\"f:terminationGracePeriodSeconds\":{},\"f:volumes\":{\".\":{},\"k:{\\\"name\\\":\\\"provider-config-volume\\\"}\":{\".\":{},\"f:configMap\":{\".\":{},\"f:defaultMode\":{},\"f:name\":{}},\"f:name\":{}},\"k:{\\\"name\\\":\\\"secret-config-volume\\\"}\":{\".\":{},\"f:name\":{},\"f:secret\":{\".\":{},\"f:defaultMode\":{},\"f:optional\":{},\"f:secretName\":{}}}}}}}},\"manager\":\"rollouts-controller\",\"operation\":\"Update\",\"time\":\"2022-11-10T05:41:00Z\"},{\"apiVersion\":\"batch/v1\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:status\":{\"f:conditions\":{}}},\"manager\":\"Argo-MetricProvider-Job\",\"operation\":\"Update\",\"subresource\":\"status\",\"time\":\"2022-11-10T05:41:03Z\"}],\"name\":\"37700755-22b7-482f-b9d9-c26518d06140.junaid-job-analysis.1\",\"namespace\":\"argocd\",\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"blockOwnerDeletion\":true,\"controller\":true,\"kind\":\"AnalysisRun\",\"name\":\"junaid-on-job-db465c966-2-2\",\"uid\":\"37700755-22b7-482f-b9d9-c26518d06140\"}],\"resourceVersion\":\"51067114\",\"uid\":\"2b00f511-6266-428d-9ba7-956dc18e0d27\"},\"spec\":{\"backoffLimit\":2,\"completionMode\":\"NonIndexed\",\"completions\":1,\"parallelism\":1,\"selector\":{\"matchLabels\":{\"controller-uid\":\"2b00f511-6266-428d-9ba7-956dc18e0d27\"}},\"suspend\":false,\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"controller-uid\":\"2b00f511-6266-428d-9ba7-956dc18e0d27\",\"job-name\":\"37700755-22b7-482f-b9d9-c26518d06140.junaid-job-analysis.1\"}},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"MY_POD_NAME\",\"valueFrom\":{\"fieldRef\":{\"apiVersion\":\"v1\",\"fieldPath\":\"metadata.name\"}}},{\"name\":\"STABLE_POD_HASH\",\"value\":\"5d4bfb9547\"},{\"name\":\"LATEST_POD_HASH\",\"value\":\"db465c966\"},{\"name\":\"PROVIDER_CONFIG\",\"value\":\"providerConfig\"}],\"image\":\"genos1998/argojob:v7\",\"imagePullPolicy\":\"IfNotPresent\",\"name\":\"junaid-job-analysis\",\"resources\":{},\"terminationMessagePath\":\"/dev/termination-log\",\"terminationMessagePolicy\":\"File\",\"volumeMounts\":[{\"mountPath\":\"/etc/config/provider\",\"name\":\"provider-config-volume\"},{\"mountPath\":\"/etc/config/secrets\",\"name\":\"secret-config-volume\",\"readOnly\":true}]}],\"dnsPolicy\":\"ClusterFirst\",\"restartPolicy\":\"Never\",\"schedulerName\":\"default-scheduler\",\"securityContext\":{},\"serviceAccount\":\"job-acc\",\"serviceAccountName\":\"job-acc\",\"terminationGracePeriodSeconds\":30,\"volumes\":[{\"configMap\":{\"defaultMode\":420,\"name\":\"opsmx-provider-config\"},\"name\":\"provider-config-volume\"},{\"name\":\"secret-config-volume\",\"secret\":{\"defaultMode\":420,\"optional\":false,\"secretName\":\"opsmx-profile-v5\"}}]}}},\"status\":{\"active\":1,\"conditions\":[{\"lastProbeTime\":\"2022-11-10T05:44:35Z\",\"lastTransitionTime\":null,\"message\":\"The analysis was Failed\",\"status\":\"True\",\"type\":\"Failed\"},{\"lastProbeTime\":\"2022-11-10T05:44:35Z\",\"lastTransitionTime\":null,\"message\":\"Canary ID: 1815\\nReport URL: https://isd-dev.argo-dev.opsmx.net/ui/argocontext/final-job/130/argoagent06/argocd6/deploymentverification/final-job/1815\\nGate URL: https://isd-dev.argo-dev.opsmx.net/\\nScore: 0\",\"status\":\"True\",\"type\":\"OpsmxAnalysis\"}],\"startTime\":\"2022-11-10T05:41:00Z\"}}"
            };
            if (data.manifest.includes('message')) {
              let a = JSON.parse(data.manifest);
              console.log(a);
              console.log(a.status.conditions[a.status.conditions.length - 1].message);
              if (a.status?.conditions[a.status.conditions.length - 1]?.message) {
                let stringValue = a.status?.conditions[a.status.conditions.length - 1]?.message.split(/\n/)[1];
                var reportURL = stringValue.substring(stringValue.indexOf(':') + 1).trim();
                console.log(reportURL);
                setURL(reportURL);
                setLoading(false)
                //window.open(reportURL, '_blank');
              }
            }
          }).catch(err => {
          });
      }
      React.useEffect(() => {
        { LoadApiCalls(props) }
      }, []);

      return (
        <WaitFor loading={loading}>
        <div style={{ margin: '1em 0', display: 'flex', alignItems: 'center', whiteSpace: 'nowrap' }}>
          <div style={{ width: '100%', display: 'flex', alignItems: 'center', height: '2em', minWidth: 0 }}>
            NEW WIDGET- IFRAME {getURL}
            &nbsp;&nbsp;&nbsp;&nbsp;
            <ActionButton
              action={() => props.clickback()}
              label='Back'
              icon='fa-undo-alt'
              style={{ fontSize: '13px' }}
            />
          </div>
        </div>
        </WaitFor>
      );
};