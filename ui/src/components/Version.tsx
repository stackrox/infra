import React from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';

import { VersionServiceApi, V1Version } from 'generated/client';
import configuration from 'client/configuration';
import useApiQuery from 'client/useApiQuery';

const api = new VersionServiceApi(configuration);

const dataFetcher = (): AxiosPromise<V1Version> => api.getVersion();

export default function Version(): JSX.Element {
  const { loading, error, data } = useApiQuery(dataFetcher);

  if (loading) return <p>Loading...</p>;
  if (error || !data) return <p>Error: {error?.message || 'unexpected server response'}</p>;

  return (
    <div style={{ fontSize: 'small' }}>
      <p>Version: {data.Version}</p>
      <p>Build Date: {moment(data.BuildDate).format('LLL')}</p>
    </div>
  );
}
