import { WaitFor } from 'argo-ui/v2';
import * as React from 'react';
import './rollout.scss';
import '../pods/pods.scss';


export const ReportsWidget = (props: {  clickback: any; reportsInput: {}}) => {
    const [getURL, setURL] = React.useState('');
    const [analysisName, setAnalysisName] = React.useState('');
    const [validUrl, setValidUrl] = React.useState(true);
    const [loading, setLoading] = React.useState(true);
    const isValidUrl = (props: string) => {
      try {
        new URL(props);
        return true;
      } catch (err) {
        return false;
      }
    }
    const LoadApiCalls = (props: any) => {
      // console.log('1stapi',props);
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
              // console.log(b);
              if (b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']) {
                fetchEndpointURL(props.reportsInput.appName, props.reportsInput.resourceName, props.reportsInput.nameSpace, props.reportsInput.version, b.status?.metricResults[b.status.metricResults.length - 1]?.measurements[b.status.metricResults.length - 1]?.metadata['job-name']);
              }
            }else{
              setValidUrl(false);
              setLoading(false);
            }
          }).catch(err => {
            setValidUrl(false);
            setLoading(false);
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
              // console.log(a);
              var indexValue = a.status.hasOwnProperty('succeeded')? a.status.conditions.length - 2: a.status.conditions.length - 1;
              // console.log(a.status.conditions[indexValue].message);
              if (a.status?.conditions[indexValue]?.message) {
                let stringValue2 = a.status?.conditions[indexValue]?.message.split(/\n/)[4];
                let stringValue = a.status?.conditions[indexValue]?.message.split(/\n/)[3];
                // let stringValue1 = a.status?.conditions[indexValue]?.message.split(/\n/)[1];
                // var user =  stringValue1.substring(stringValue1.indexOf(':') + 1).trim();
                if(stringValue.split(':')[0].trim() == "reportURL"){
                  var reportId =  stringValue2.substring(stringValue2.indexOf(':') + 1).trim();
                  if(reportId){
                    var reportURL = stringValue.substring(stringValue.indexOf(':') + 1).trim() + `&p=${reportId}`;
                  }
                // console.log(user);
                  if(isValidUrl(reportURL)){
                    setValidUrl(true);
                    setURL(reportURL);
                    setLoading(false);
                    // let b = setInterval(() => {
                //   let reportPage = document.getElementById("reportPage") as HTMLIFrameElement;
                //   if(reportPage){
                //     reportPage.contentWindow.postMessage(user,"*");
                //     // console.log("IFRAME MSG ", reportPage);
                //   }
                // }, 100);
                // window.addEventListener('message', function(event) {
                //   if(typeof event.data == 'string' && event.origin != window.location.origin) {
                //     // console.log(event.data); // Message received from parent
                //     clearInterval(b);
                //   }
                // });
                //window.open(reportURL, '_blank');
              
                  }else{
                    setValidUrl(false);
                    setLoading(false);
                  }
                }else{
                  setValidUrl(false);
                  setLoading(false);
                }
              }
              
            }else{
              setValidUrl(false);
              setLoading(false);
            }
          }).catch(err => {
            setValidUrl(false);
            setLoading(false);
          });
      }
      React.useEffect(() => {
        { LoadApiCalls(props) }
      }, []);

      return (
      <WaitFor loading={loading}>
        <div style={{ margin: '1em', width: '100%', height: '100%' }}>
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
            {validUrl && <iframe src={getURL} width="100%" height="90%"></iframe>}
            {!validUrl && 
            <div style={{width: '100%', height: '100%', backgroundColor:'#dae1eb'}}>
              <div className="reports__error">
                <div className="empty-state">
                  <div className="empty-state__icon">
                    <i className="fa fa-file"></i>
                  </div>
                  <h4 style={{color:'#155362'}}>Analysis not performed or Report not available</h4>
                  <h5></h5>
                </div>
              </div>
            </div>}
          </div>
      </WaitFor>
      );
};
