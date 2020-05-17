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

export default function UserServiceAccountToken(): ReactElement {
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
    <div>
      <div className="text-xl mb-8">
        <p className="my-2">
          After downloading the file, you can install it anywhere in your <code>$PATH</code>. For
          example, you may put it in your Go executable directory.
        </p>
        <p className="my-2">
          Here are the commands to move the file, allow it to execute on a Mac, confirm its
          location, and help you learn about its features.
        </p>
        <pre className="border border-base-400 p-4 text-lg whitespace-pre-wrap">
          $ install ~/Downloads/infractl-darwin-amd64 $GOPATH/bin/infractl
          <br />$ xattr -c $GOPATH/bin/infractl
          <br />$ which infractl
          <br />$ infractl help
        </pre>
      </div>

      <h3 className="text-3xl mb-2">Authenticating with infractl</h3>
      <div className="flex items-center">
        <p className="text-xl">Run the following in a terminal to authenticate infractl for use:</p>
        <button
          type="button"
          aria-label="Copy to clipboard"
          onClick={clipboard.copy}
          className="ml-2"
        >
          <Tooltip content={<TooltipOverlay>Copy to clipboard</TooltipOverlay>}>
            <div className="flex items-center">
              <Clipboard size={16} />
              {clipboard.copied && <span className="ml-2 text-success-700">Copied!</span>}
            </div>
          </Tooltip>
        </button>
      </div>
      <textarea
        rows={6}
        value={exportCommand}
        className="mt-4 w-full bg-base-100 p-1 rounded border-2 border-base-300 hover:border-base-400 font-600 leading-normal outline-none"
        readOnly
        ref={clipboard.target}
      />
    </div>
  );
}
