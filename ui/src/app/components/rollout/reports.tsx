import { ActionButton, WaitFor } from 'argo-ui/v2';
import * as React from 'react';
import './rollout.scss';
import '../pods/pods.scss';
// import LoadingSpinner from "./LoadingSpinner";


export const ReportsWidget = (props: {  clickback: any; reportsInput: {}}) => {
    const [getURL, setURL] = React.useState('');
    const [loading, setLoading] = React.useState(true);
    const LoadApiCalls = (props: any) => {
      console.log('1stapi',props);
        setLoading(true);
        let url2 = '/api/v1/applications/' + props.reportsInput.appName + '/resource?name=' + props.reportsInput.resourceName + '&appNamespace=' + props.reportsInput.nameSpace + '&namespace=' + props.reportsInput.nameSpace + '&resourceName=' + props.reportsInput.resourceName + '&version=' + props.reportsInput.version + '&kind=AnalysisRun&group=argoproj.io';
        fetch(url2)
          .then(response => {
            return response.json()
          })
          .then((data: any) => {
            if (data.manifest.includes('job-name')) {
              let b = JSON.parse(data.manifest);
              console.log(b);
              if (b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']) {
                fetchEndpointURL(props.reportsInput.appName, props.reportsInput.resourceName, props.reportsInput.nameSpace, props.reportsInput.version, b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']);
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
                let iframeWin = document.getElementById("iframeWin") as HTMLIFrameElement;
                if(iframeWin){
                  iframeWin.contentWindow.postMessage("MESSAGE", "admin");
                  console.log("IFRAME MSG ", iframeWin)
                }
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
        <div style={{ margin: '1em', width: '100%', height: '100%' }}>
        <ActionButton
              action={() => props.clickback()}
              label='Back'
              icon='fa-undo-alt'
              style={{ fontSize: '13px', width: '7%', marginBottom: '1em', marginLeft: 'auto' }}
            />
          <div style={{ width: '100%', alignItems: 'center', height: '100%' }}>
          <iframe src={getURL} width="100%" height="90%" id="iframeWin"></iframe>
           
          </div>
        </div>
        </WaitFor>
      );
};