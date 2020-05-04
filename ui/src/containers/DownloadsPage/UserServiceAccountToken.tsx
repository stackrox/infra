import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import { useClipboard } from 'use-clipboard-copy';

import { V1TokenResponse, UserServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import Tooltip from 'components/Tooltip';
import TooltipOverlay from 'components/TooltipOverlay';
import { AlertCircle, Clipboard } from 'react-feather';
import { ClipLoader } from 'react-spinners';

const userService = new UserServiceApi(configuration);

const fetchToken = (): AxiosPromise<V1TokenResponse> => userService.token({});

type Props = {
  className?: string;
};

export default function UserServiceAccountToken({ className = '' }: Props): ReactElement {
  const { loading, error, data } = useApiQuery(fetchToken);
  const clipboard = useClipboard({
    copiedTimeout: 800, // duration in milliseconds to show 'successfully copied' feedback
  });

  if (loading) {
    return (
      <div className="inline-flex items-center">
        <ClipLoader size={16} color="currentColor" />
        <span className="ml-2">Loading service account token...</span>
      </div>
    );
  }

  if (error || !data?.Token) {
    return (
      <Tooltip content={<TooltipOverlay>{error?.message || 'Unknown error'}</TooltipOverlay>}>
        <div className="inline-flex items-center">
          <AlertCircle size={16} />
          <span className="ml-2">
            Unexpected error occurred while loading service account token
          </span>
        </div>
      </Tooltip>
    );
  }

  const exportCommand = `export INFRA_TOKEN="${data.Token}"`;

  return (
    <div className={className}>
      <div className="flex items-center">
        <span className="text-xl">
          Run the following in a terminal to configure infractl for use:
        </span>
        <button type="button" onClick={clipboard.copy} className="ml-2">
          <Tooltip content={<TooltipOverlay>Copy to clipboard</TooltipOverlay>}>
            <div className="flex items-center">
              <Clipboard size={16} />
              {clipboard.copied && <span className="ml-2 text-success-700">Copied!</span>}
            </div>
          </Tooltip>
        </button>
      </div>
      <input
        type="text"
        value={exportCommand}
        className="mt-4 w-full bg-base-100 p-1 rounded border-2 border-base-300 hover:border-base-400 font-600 leading-normal outline-none"
        readOnly
        ref={clipboard.target}
      />
    </div>
  );
}
