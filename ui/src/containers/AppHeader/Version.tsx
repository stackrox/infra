import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import { Info, AlertCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';

import { VersionServiceApi, V1Version } from 'generated/client';
import configuration from 'client/configuration';
import useApiQuery from 'client/useApiQuery';

const versionService = new VersionServiceApi(configuration);

const fetchVersion = (): AxiosPromise<V1Version> => versionService.getVersion();

function VersionContent(props: { icon: ReactElement; text: string }): ReactElement {
  const { icon, text } = props;
  return (
    <div className="flex items-center">
      {icon}
      <span className="ml-1 text-2xs whitespace-nowrap">{text}</span>
    </div>
  );
}

export default function Version(): ReactElement {
  const { loading, error, data } = useApiQuery(fetchVersion);

  if (loading)
    return (
      <VersionContent
        icon={<ClipLoader loading size={16} color="currentColor" />}
        text="Loading..."
      />
    );

  if (error || !data?.Version)
    return <VersionContent icon={<AlertCircle size={16} />} text="Error occurred" />;

  return <VersionContent icon={<Info size={16} />} text={data.Version} />;
}
