import React, { ReactElement } from 'react';
import {
  ClipboardCopy,
  ClipboardCopyVariant,
  CodeBlock,
  CodeBlockCode,
  Divider,
  Flex,
  Title,
} from '@patternfly/react-core';
import { AxiosPromise } from 'axios';
import { AlertCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';

import { V1TokenResponse, UserServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';

const userService = new UserServiceApi(configuration);

const fetchToken = (): AxiosPromise<V1TokenResponse> => userService.token({});

export default function UserServiceAccountToken(): ReactElement {
  const { loading, error, data } = useApiQuery(fetchToken);

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
      <div className="inline-flex items-center">
        <AlertCircle size={16} />
        <span className="ml-2">Unexpected error occurred while loading service account token</span>
      </div>
    );
  }

  const exportCommand = `export INFRA_TOKEN="${data.Token}"`;

  return (
    <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsMd' }}>
      <p>
        After downloading the file, you can install it anywhere in your <code>$PATH</code>. For
        example, you may put it in your Go executable directory.
      </p>
      <p>
        Here are the commands to move the file, allow it to execute on an Intel-based Mac, confirm
        its location, and help you learn about its features.
      </p>
      <CodeBlock className="pf-v6-u-w-50-on-xl">
        <CodeBlockCode>
          $ install ~/Downloads/infractl-darwin-amd64 $GOPATH/bin/infractl
          <br />$ xattr -c $GOPATH/bin/infractl
          <br />$ which infractl
          <br />$ infractl help
        </CodeBlockCode>
      </CodeBlock>

      <Divider component="div" />

      <Title headingLevel="h3">Authenticating with infractl</Title>
      <p>Run the following in a terminal to authenticate infractl for use:</p>
      <ClipboardCopy
        className="pf-v6-u-w-50-on-xl"
        isReadOnly
        hoverTip="Copy"
        clickTip="Copied"
        variant={ClipboardCopyVariant.expansion}
      >
        {exportCommand}
      </ClipboardCopy>
    </Flex>
  );
}
