import {ActionButton, EffectDiv, formatTimestamp, InfoItemProps, InfoItemRow, ThemeDiv, Tooltip} from 'argo-ui/v2';
import * as React from 'react';
import {RolloutAnalysisRunInfo, RolloutExperimentInfo, RolloutReplicaSetInfo} from '../../../models/rollout/generated';
import {IconForTag} from '../../shared/utils/utils';
import {PodWidget, ReplicaSets} from '../pods/pods';
import {ImageInfo, parseImages} from './rollout';
import './rollout.scss';
import '../pods/pods.scss';

export interface Revision {
    number: string;
    replicaSets: RolloutReplicaSetInfo[];
    experiments: RolloutExperimentInfo[];
    analysisRuns: RolloutAnalysisRunInfo[];
}

const handleClick = (applicationName: String, resouceName: String, nameSpace: String, version: String) => {
    let url1 = '/api/v1/applications/' + applicationName + '/resource?name=' + resouceName + '&appNamespace=' + nameSpace + '&namespace=' + nameSpace + '&resourceName=' + resouceName + '&version=' + version + '&kind=AnalysisRun&group=argoproj.io';
    fetch(url1)
        .then(response => {
            // if (response.status > 399) {
            //   throw new Error("No metrics");
            // }
            return response.json()
      })
      .then((data: any) => {
        if(data.manifest.includes('job-name')){
            let b = JSON.parse(data.manifest);
            if(b.status?.metricResults[b.status.metricResults.length-1]?.measurements[b.status.metricResults.length-1]?.metadata['job-name']){
                fetchEndpointURL(applicationName,resouceName,nameSpace,version,b.status?.metricResults[b.status.metricResults.length-1]?.measurements[b.status.metricResults.length-1]?.metadata['job-name']);
            }
        }
      }).catch(err => {
       
      });
};

const fetchEndpointURL = (applicationName:String,resouceName:String,nameSpace:String,version:String,jobName:String) => {
    let url2 = '/api/v1/applications/'+applicationName+'/resource?name='+jobName+'&appNamespace='+nameSpace+'&namespace='+nameSpace+'&resourceName='+jobName+'&version=v1&kind=Job&group=batch'
    fetch(url2)
    .then(response => {
        return response.json()
      })
      .then((data: any) => {
        if(data.manifest.includes('message')){
            let a = JSON.parse(data.manifest);
            console.log(a.status.conditions[a.status.conditions.length-1].message);
            if(a.status?.conditions[a.status.conditions.length-1]?.message){
                let stringValue = a.status?.conditions[a.status.conditions.length-1]?.message.split(/\n/)[1];
                var reportURL = stringValue.substring(stringValue.indexOf(':') + 1).trim();
                console.log(reportURL);
                window.open(reportURL, '_blank');
            }
        }
      }).catch(err => {
        console.error('res.data', err)
      });
};

const ImageItems = (props: { images: ImageInfo[] }) => {
    return (
        <div>
            {props.images.map((img) => {
                let imageItems = img?.tags?.map((t) => {
                    return {content: t, icon: IconForTag(t)} as InfoItemProps;
                });
                if (imageItems.length === 0) {
                    imageItems = [];
                }
                return <InfoItemRow key={img.image} label={<ThemeDiv className={`image image--${img.color || 'unknown'}`}>{img.image}</ThemeDiv>} items={imageItems} />;
            })}
        </div>
    );
};

interface RevisionWidgetProps {
    revision: Revision;
    initCollapsed?: boolean;
    rollback?: (revision: number) => void;
    current: boolean;
    message: String;
    appName: String;
}

export const RevisionWidget = (props: RevisionWidgetProps) => {
    const {revision, initCollapsed} = props;
    const [collapsed, setCollapsed] = React.useState(initCollapsed);
    const icon = collapsed ? 'fa-chevron-circle-down' : 'fa-chevron-circle-up';
    const images = parseImages(revision.replicaSets);
    return (
        <EffectDiv key={revision.number} className='revision'>
            <ThemeDiv className='revision__header'>
                Revision {revision.number}
                <div style={{marginLeft: 'auto', display: 'flex', alignItems: 'center'}}>
                    {!props.current && props.rollback && (
                        <ActionButton
                            action={() => props.rollback(Number(revision.number))}
                            label='ROLLBACK'
                            icon='fa-undo-alt'
                            style={{fontSize: '13px'}}
                            indicateLoading
                            shouldConfirm
                        />
                    )}
                    <ThemeDiv className='revision__header__button' onClick={() => setCollapsed(!collapsed)}>
                        <i className={`fa ${icon}`} />
                    </ThemeDiv>
                </div>
            </ThemeDiv>
            <ThemeDiv className='revision__images'>
                <ImageItems images={images} />
            </ThemeDiv>

            {!collapsed && (
                <React.Fragment>
                    <ReplicaSets replicaSets={revision.replicaSets} />
                    {(revision.analysisRuns || []).length > 0 && (
                        <React.Fragment>
                            <div style={{marginTop: '1em'}}>
                                <AnalysisRunWidget analysisRuns={revision.analysisRuns} appName={props.appName} />
                            </div>
                        </React.Fragment>
                    )}
                </React.Fragment>
            )}
        </EffectDiv>
    );
};

const AnalysisRunWidget = (props: { analysisRuns: RolloutAnalysisRunInfo[], appName?: String }) => {
    const {analysisRuns} = props;
    const [selection, setSelection] = React.useState<RolloutAnalysisRunInfo>(null);

    return (
        <ThemeDiv className='analysis'>
            <div className='analysis-header'>Analysis Runs</div>
            <div className='analysis__runs'>
                {analysisRuns.map((ar) => {
                    let temp = ar.objectMeta.name.split('-');
                    let len = temp.length;
                    let resourceName = ar.objectMeta.name;
                    let namespace = ar.objectMeta.namespace;
                    let version = ar.objectMeta.resourceVersion;
                    return (
                        <Tooltip
                            key={ar.objectMeta?.name}
                            content={
                                <React.Fragment>
                                    <div>
                                        <b>Name:</b> {ar.objectMeta.name}
                                    </div>
                                    <div>
                                        <b>Created at: </b>
                                        {formatTimestamp(JSON.stringify(ar.objectMeta?.creationTimestamp))}
                                    </div>
                                    <div>
                                        <b>Status: </b>
                                        {ar.status}
                                    </div>
                                </React.Fragment>
                            }>
                            <div
                                className={`analysis__runs-action ${
                                    ar.status === 'Running' ? 'analysis--pending' : ar.status === 'Successful' ? 'analysis--success' : 'analysis--failure'
                                }`}>
                                <ActionButton
                                    action={() => {
                                        (selection?.objectMeta.name === ar.objectMeta.name ? setSelection(null) : setSelection(ar));
                                        if (props?.appName) {
                                            handleClick(props.appName, resourceName, namespace, version)
                                        }
                                    }}
                                    label={`Analysis ${temp[len - 2] + '-' + temp[len - 1]}`}
                                />
                            </div>
                        </Tooltip>
                    );
                })}
            </div>

            {selection && (
                <React.Fragment key={selection.objectMeta?.name}>
                    <div style={{marginTop: 5}}>
                        {selection.objectMeta?.name}
                        <i className={`fa ${selection.status === 'Successful' ? 'fa-check-circle analysis--success' : 'fa-times-circle analysis--failure'}`} />
                        {/* <a href="google.com" >View Report</a> */}
                    </div>
                    {selection?.jobs && (
                        <div className='analysis__run__jobs'>
                            <div className='analysis__run__jobs-list'>
                                {selection.jobs.map((job) => {
                                    return (
                                        <PodWidget
                                            key={job.objectMeta?.name}
                                            name={job.objectMeta.name}
                                            status={job.status}
                                            tooltip={
                                                <div>
                                                    <div>job-name: {job.objectMeta?.name}</div>
                                                    <div>StartedAt: {formatTimestamp(JSON.stringify(job.startedAt))}</div>
                                                    <div>Status: {job.status}</div>
                                                    <div>MetricName: {job.metricName}</div>
                                                </div>
                                            }
                                            customIcon='fa-chart-bar'
                                        />
                                    );
                                })}
                            </div>
                            <Tooltip
                                content={selection?.metrics
                                    .filter((metric) => metric.name === selection.jobs[0].metricName)
                                    .map((metric) => {
                                        return (
                                            <React.Fragment key={metric.name}>
                                                {metric?.name && (
                                                    <div>
                                                        <b>MetricName:</b> {metric.name}
                                                    </div>
                                                )}
                                                {metric?.successCondition && (
                                                    <div>
                                                        <b>SuccessCondition: </b>
                                                        {metric.successCondition}
                                                    </div>
                                                )}
                                                {metric?.failureLimit && (
                                                    <div>
                                                        <b>FailureLimit:</b> {metric.failureLimit}
                                                    </div>
                                                )}
                                                {metric?.inconclusiveLimit && (
                                                    <div>
                                                        <b>InconclusiveLimit: </b>
                                                        {metric.inconclusiveLimit}
                                                    </div>
                                                )}
                                                {metric?.count && (
                                                    <div>
                                                        <b>Count: </b>
                                                        {metric.count}
                                                    </div>
                                                )}
                                            </React.Fragment>
                                        );
                                    })}>
                                <i className='fa fa-info-circle analysis__run__jobs-info' />
                            </Tooltip>
                        </div>
                    )}
                    {selection?.nonJobInfo && (
                        <div className='analysis__run__jobs'>
                            <div className='analysis__run__jobs-list'>
                                {selection.nonJobInfo.map((nonJob) => {
                                    return (
                                        <PodWidget
                                            key={new Date(nonJob.startedAt.seconds).getTime()}
                                            name={nonJob.value}
                                            status={nonJob.status}
                                            tooltip={
                                                <div>
                                                    <pre>Value: {JSON.stringify(JSON.parse(nonJob.value), null, 2)}</pre>
                                                    <div>StartedAt: {formatTimestamp(JSON.stringify(nonJob.startedAt))}</div>
                                                    <div>Status: {nonJob.status}</div>
                                                    <div>MetricName: {nonJob.metricName}</div>
                                                </div>
                                            }
                                            customIcon='fa-chart-bar'
                                        />
                                    );
                                })}
                            </div>
                            <Tooltip
                                content={selection?.metrics
                                    .filter((metric) => metric.name === selection.nonJobInfo[0].metricName)
                                    .map((metric) => {
                                        return (
                                            <React.Fragment key={metric.name}>
                                                {metric?.name && (
                                                    <div>
                                                        <b>MetricName:</b> {metric.name}
                                                    </div>
                                                )}
                                                {metric?.successCondition && (
                                                    <div>
                                                        <b>SuccessCondition: </b>
                                                        {metric.successCondition}
                                                    </div>
                                                )}
                                                {metric?.failureLimit && (
                                                    <div>
                                                        <b>FailureLimit:</b> {metric.failureLimit}
                                                    </div>
                                                )}
                                                {metric?.inconclusiveLimit && (
                                                    <div>
                                                        <b>InconclusiveLimit: </b>
                                                        {metric.inconclusiveLimit}
                                                    </div>
                                                )}
                                                {metric?.count && (
                                                    <div>
                                                        <b>Count: </b>
                                                        {metric.count}
                                                    </div>
                                                )}
                                            </React.Fragment>
                                        );
                                    })}>
                                <i className='fa fa-info-circle analysis__run__jobs-info' />
                            </Tooltip>
                        </div>
                    )}
                </React.Fragment>
            )}
        </ThemeDiv>
    );
};
