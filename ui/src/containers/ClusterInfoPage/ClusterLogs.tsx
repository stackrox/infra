import React, { useCallback, ReactElement } from 'react';

import { ClusterServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  clusterId: string;
};

export default function ClusterLogs({ clusterId }: Props): ReactElement {
  const fetchClusterLogs = useCallback(() => clusterService.logs(clusterId), [clusterId]);
  const { loading, error, data } = useApiQuery(fetchClusterLogs, { pollInterval: 10000 });

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Logs) {
    return <FullPageError message={error?.message || 'No logs found'} />;
  }

  // combine all log entries into a single log with sections split same way infractl does
  const logs = data.Logs.map((logEntry) => {
    if (!logEntry.Name || !logEntry.Body) return '';

    const logEntryHeaderBorder = '-'.repeat(logEntry.Name.length);
    const logText = atob(logEntry.Body);
    return `${logEntry.Name}\n${logEntryHeaderBorder}\n${logEntry.Message || ''}\n${logText}`;
  }).join('\n\n');

  return (
    <pre className="font-300 whitespace-pre-wrap">
      <code className="h-full pb-5">{logs}</code>
    </pre>
  );
}
