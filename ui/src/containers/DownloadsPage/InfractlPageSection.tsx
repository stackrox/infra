import React, { ReactElement } from 'react';
import { Button, Flex, PageSection, Title } from '@patternfly/react-core';

import { DownloadIcon } from '@patternfly/react-icons';
import UserServiceAccountToken from './UserServiceAccountToken';

export default function InfractlPageSection(): ReactElement {
  const infractlDownloads = {
    'Download for Intel Mac': '/downloads/infractl-darwin-amd64',
    'Download for M1 Mac': '/downloads/infractl-darwin-arm64',
    'Download for Linux': '/downloads/infractl-linux-amd64',
  };
  const infractlLinks = Object.entries(infractlDownloads).map(([label, value]) => (
    <Button
      component="a"
      variant="secondary"
      key={value}
      href={value}
      icon={<DownloadIcon />}
      download
    >
      {label}
    </Button>
  ));

  return (
    <PageSection>
      <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsMd' }}>
        <Title headingLevel="h1">Infractl (CLI)</Title>
        <Flex spaceItems={{ default: 'spaceItemsMd' }}>{infractlLinks}</Flex>
        <UserServiceAccountToken />
      </Flex>
    </PageSection>
  );
}
