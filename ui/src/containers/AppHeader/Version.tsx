import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';
import { Info, AlertCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';
import { Tooltip, TooltipOverlay } from '@stackrox/ui-components';

import { VersionServiceApi, V1Version } from 'generated/client';
import configuration from 'client/configuration';
import useApiQuery from 'client/useApiQuery';

const versionService = new VersionServiceApi(configuration);

const fetchVersion = (): AxiosPromise<V1Version> => versionService.getVersion();

function TooltipContentRow(props: { label: string; value?: string }): ReactElement {
  const { label, value = 'unknown' } = props;
  return (
    <div>
      <span className="font-700">{label}: </span>
      <span>{value}</span>
    </div>
  );
}

function TooltipContent(props: { data: V1Version }): ReactElement {
  const { data } = props;
  return (
    <TooltipOverlay extraClassName="text-left">
      <TooltipContentRow label="Build Date" value={moment(data.BuildDate).format('LLL')} />
      <TooltipContentRow label="Git Commit" value={data.GitCommit} />
      <TooltipContentRow label="Workflow" value={data.Workflow} />
      <TooltipContentRow label="Go Version" value={data.GoVersion} />
      <TooltipContentRow label="Platform" value={data.Platform} />
    </TooltipOverlay>
  );
}

function VersionContent(props: {
  tooltip: ReactElement;
  icon: ReactElement;
  text: string;
}): ReactElement {
  const { tooltip, icon, text } = props;
  return (
    <Tooltip content={tooltip}>
      <div className="flex items-center">
        {icon}
        <span className="ml-1 text-2xs whitespace-no-wrap">{text}</span>
      </div>
    </Tooltip>
  );
}

export default function Version(): ReactElement {
  const { loading, error, data } = useApiQuery(fetchVersion);

  if (loading)
    return (
      <VersionContent
        tooltip={<TooltipOverlay>Loading server version data...</TooltipOverlay>}
        icon={<ClipLoader loading size={16} color="currentColor" />}
        text="Loading..."
      />
    );

  if (error || !data?.Version)
    return (
      <VersionContent
        tooltip={<TooltipOverlay>{error?.message || 'Unexpected server response'}</TooltipOverlay>}
        icon={<AlertCircle size={16} />}
        text="Error occurred"
      />
    );

  return (
    <VersionContent
      tooltip={<TooltipContent data={data} />}
      icon={<Info size={16} />}
      text={data.Version}
    />
  );
}
