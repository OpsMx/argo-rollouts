import { WaitFor } from 'argo-ui/v2';
import * as React from 'react';
import './rollout.scss';
import '../pods/pods.scss';


export const ReportsWidget = (props: {  clickback: any; reportsInput: {}}) => {
    const [getURL, setURL] = React.useState('');
    const [analysisName, setAnalysisName] = React.useState('');
    const [loading, setLoading] = React.useState(true);
    const LoadApiCalls = (props: any) => {
      console.log('1stapi',props);
        setLoading(true);
        setAnalysisName(props?.reportsInput?.analysisName);
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
              var indexValue = a.status.hasOwnProperty('succeeded')? a.status.conditions.length - 2: a.status.conditions.length - 1;
              console.log(a.status.conditions[indexValue].message);
              if (a.status?.conditions[indexValue]?.message) {
                let stringValue = a.status?.conditions[indexValue]?.message.split(/\n/)[1];
                var reportURL = stringValue.substring(stringValue.indexOf(':') + 1).trim();
                console.log(reportURL);
                setURL(reportURL);
                setLoading(false);
                let b = setInterval(() => {
                  let reportPage = document.getElementById("reportPage") as HTMLIFrameElement;
                  if(reportPage){
                    reportPage.contentWindow.postMessage("admin","*");
                    console.log("IFRAME MSG ", reportPage);
                  }
                }, 100);
                window.addEventListener('message', function(event) {
                  if(typeof event.data == 'string' && event.origin != window.location.origin) {
                    console.log(event.data); // Message received from parent
                    clearInterval(b);
                  }
                });
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
          <div>
            <div className='bc-element bc-first' style={{ left: '0px' }} onClick={() => props.clickback()}>
              <div className='bc-text bc-text-first addPointer'>Back to Dashboard</div>
              <div className='bc-arrow' style={{ zIndex: 2 }}></div>
            </div>
            <div className='bc-element' style={{ left: '-5px' }}>
              <div className='bc-before-arrow bc-hefore-arrow-last'>
                <div className='bc-arrow' style={{ borderLeft: '10px solid white' }}></div>
              </div>
              <div className='bc-text bc-text-last'>{analysisName}
              </div>
              <div className='bc-arrow bc-arrow-last' style={{ zIndex: 2 }}></div>
            </div>
            <div style={{ clear: 'both' }}></div>
          </div>
          <div style={{ width: '100%', alignItems: 'center', height: '100%' }}>
            <iframe src={getURL} width="100%" height="90%" id="reportPage"></iframe>
          </div>
        </div>
        </WaitFor>
      );
};