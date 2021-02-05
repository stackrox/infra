import React, { ReactElement } from 'react';
import { useClipboard } from 'use-clipboard-copy';
import { Tooltip, TooltipOverlay } from '@stackrox/ui-components';
import { Clipboard } from 'react-feather';

type Props = {
  connect: string;
};

export default function ClusterConnect({ connect }: Props): ReactElement {
  const clipboard = useClipboard({
    copiedTimeout: 800, // duration in milliseconds to show 'successfully copied' feedback
  });

  const stripComments = connect
    .split('\n')
    .filter((line) => !/^#/.test(line))
    .join('\n');

  return (
    <span className="flex content-start text-base normal-case">
      Connect: <span className="ml-2 font-mono">{stripComments}</span>{' '}
      <button
        type="button"
        aria-label="Copy to clipboard"
        onClick={() => clipboard.copy(stripComments)}
        className="ml-2"
      >
        <Tooltip content={<TooltipOverlay>Copy to clipboard</TooltipOverlay>}>
          <div className="flex items-center">
            <Clipboard size={16} />
            {clipboard.copied && <span className="ml-2 text-success-700">Copied!</span>}
          </div>
        </Tooltip>
      </button>
    </span>
  );
}
