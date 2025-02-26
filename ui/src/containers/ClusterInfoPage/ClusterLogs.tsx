import React, { ReactElement } from 'react';

import { ClusterServiceApi, V1Log } from 'generated/client';
import configuration from 'client/configuration';
import { CodeBlock, CodeBlockCode } from '@patternfly/react-core';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  logs: V1Log[];
};

export default function ClusterLogs({ logs }: Props): ReactElement {
  // combine all log entries into a single log with sections split same way infractl does
  const logsJoined = logs
    .map((logEntry) => {
      if (!logEntry.Name || !logEntry.Body) return '';

      const logEntryHeaderBorder = '-'.repeat(logEntry.Name.length);
      const logText = atob(logEntry.Body);
      return `${logEntry.Name}\n${logEntryHeaderBorder}\n${logEntry.Message || ''}\n${logText}`;
    })
    .join('\n\n');

  return (
    <CodeBlock>
      <CodeBlockCode>{logsJoined}</CodeBlockCode>
    </CodeBlock>
  );
}
