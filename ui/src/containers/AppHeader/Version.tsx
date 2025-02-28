import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import { Flex, FlexItem, Spinner } from '@patternfly/react-core';
import { InfoAltIcon, InfoCircleIcon } from '@patternfly/react-icons';

import { VersionServiceApi, V1Version } from 'generated/client';
import configuration from 'client/configuration';
import useApiQuery from 'client/useApiQuery';

const versionService = new VersionServiceApi(configuration);

const fetchVersion = (): AxiosPromise<V1Version> => versionService.getVersion();

function VersionContent(props: { icon: ReactElement; text: string }): ReactElement {
  const { icon, text } = props;
  return (
    <Flex alignItems={{ default: 'alignItemsFlexStart' }}>
      <FlexItem spacer={{ default: 'spacerSm' }}>{icon}</FlexItem>
      <FlexItem>{text}</FlexItem>
    </Flex>
  );
}

export default function Version(): ReactElement {
  const { loading, error, data } = useApiQuery(fetchVersion);

  if (loading)
    return (
      <VersionContent
        icon={<Spinner size="md" aria-label="Loading version information" />}
        text="Loading..."
      />
    );

  if (error || !data?.Version)
    return (
      <VersionContent
        icon={<InfoCircleIcon color="var(--pf-t--global--icon--color--status--danger--default)" />}
        text="Could not load server version"
      />
    );

  return <VersionContent icon={<InfoAltIcon />} text={data.Version} />;
}
