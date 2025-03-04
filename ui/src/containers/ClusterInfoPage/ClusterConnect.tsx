import React, { ReactElement } from 'react';
import { ClipboardCopy, Flex } from '@patternfly/react-core';

type Props = {
  connect: string;
};

export default function ClusterConnect({ connect }: Props): ReactElement {
  const stripComments = connect
    .split('\n')
    .filter((line) => !/^#/.test(line))
    .join('\n');

  return (
    <Flex alignItems={{ default: 'alignItemsCenter' }}>
      <span>Connect:</span>
      <ClipboardCopy
        className="pf-v6-u-flex-grow-1"
        isReadOnly
        hoverTip="Copy command"
        clickTip="Command copied!"
      >
        {stripComments}
      </ClipboardCopy>
    </Flex>
  );
}
