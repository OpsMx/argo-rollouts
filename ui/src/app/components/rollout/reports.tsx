import { WaitFor } from 'argo-ui/v2';
import * as React from 'react';
import './rollout.scss';
import '../pods/pods.scss';


export const ReportsWidget = (props: {  clickback: any; reportsInput: {}}) => {
    const [getURL, setURL] = React.useState('');
    const [analysisName, setAnalysisName] = React.useState('');
    const [validUrl, setValidUrl] = React.useState(true);
    const [loading, setLoading] = React.useState(true);
    let conditionArray: any[] = [];
    let jobsList: any[] = [];
    let lastIterationJob = 0;
    const isValidUrl = (props: string) => {
      try {
        new URL(props);
        return true;
      } catch (err) {
        return false;
      }
    }
    const LoadApiCalls = (props: any) => {
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
              const newJobs: any[] = [];
              const mResults = b.status?.metricResults;
              mResults.forEach((element: any, index: number) => {
                newJobs.push(element.measurements[0]?.metadata['job-name']);
                if (b.status?.metricResults[index]?.measurements[0]?.metadata['job-name']) {
                  fetchEndpointURL(props.reportsInput.appName, props.reportsInput.resourceName, props.reportsInput.nameSpace, props.reportsInput.version, b.status?.metricResults[index]?.measurements[0]?.metadata['job-name'], index);
                }
                jobsList = newJobs;
                lastIterationJob = newJobs.length - 1;
              });
            }else{
              setValidUrl(false);
              setLoading(false);
            }
          }).catch(err => {
            setValidUrl(false);
            setLoading(false);
          });
      };

    const fetchEndpointURL = (applicationName: String, resouceName: String, nameSpace: String, version: String, jobName: String, index: number) => {
        let url3 = '/api/v1/applications/' + applicationName + '/resource?name=' + jobName + '&appNamespace=' + nameSpace + '&namespace=' + nameSpace + '&resourceName=' + jobName + '&version=v1&kind=Job&group=batch'
        fetch(url3)
          .then(response => {
            return response.json()
          })
          .then((data: any) => {
            if(data.manifest.includes('message')) {
            let a = JSON.parse(data.manifest);
            a.status.conditions.filter((conlist: { type: string; }) => conlist.type === 'OpsmxAnalysis').map((element: any) => { conditionArray = [...conditionArray, element] });
            const latest = conditionArray.reduce(function (r, a) {
              return r.lastProbeTime > a.lastProbeTime ? r : a;
            });
    
            if(jobsList.length - 1 === lastIterationJob){
              let stringValue2 = latest.message.split(/\n/)[4];
              let stringValue = latest.message.split(/\n/)[3];
              if(stringValue.split(':')[0].trim() == "reportURL"){
                var reportId = stringValue2.substring(stringValue2.indexOf(':') + 1).trim();
                if(reportId){
                  var reportURL = stringValue.substring(stringValue.indexOf(':') + 1).trim() + `&p=${reportId}`;
                }
                if(isValidUrl(reportURL)){
                  setTimeout(() => {
                    setValidUrl(true);
                    setLoading(false);
                    setURL(reportURL);
                  }, 600);
                }else{
                  setValidUrl(false);
                  setLoading(false);
                }
              } 
            }
          }else{
            if (jobsList.length - 1 === lastIterationJob) {
              setValidUrl(false);
              setLoading(false);
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
                  <h5 style={{color:'#155362'}}>Analysis not performed or Report not available</h5>
                </div>
              </div>
            </div>}
          </div>
      </WaitFor>
      );
};
