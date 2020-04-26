import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';
import { Info, AlertCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';

import { VersionServiceApi, V1Version } from 'generated/client';
import configuration from 'client/configuration';
import useApiQuery from 'client/useApiQuery';
import Tooltip from 'components/shared/Tooltip';
import TooltipOverlay from 'components/shared/TooltipOverlay';

const api = new VersionServiceApi(configuration);

const dataFetcher = (): AxiosPromise<V1Version> => api.getVersion();

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
    <TooltipOverlay className="text-left">
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
        <span className="ml-1 text-2xs">{text}</span>
      </div>
    </Tooltip>
  );
}

export default function Version(): ReactElement {
  const { loading, error, data } = useApiQuery(dataFetcher);

  if (loading)
    return (
      <VersionContent
        tooltip={<TooltipOverlay>Loading data...</TooltipOverlay>}
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
