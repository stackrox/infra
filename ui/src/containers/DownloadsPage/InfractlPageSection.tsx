import React, { ReactElement } from 'react';

import PageSection from 'components/PageSection';
import UserServiceAccountToken from './UserServiceAccountToken';

export default function InfractlPageSection(): ReactElement {
  const infractlDownloads = {
    'Download for Intel-Mac': '/downloads/infractl-darwin-amd64',
    'Download for M1-Mac': '/downloads/infractl-darwin-arm64',
    'Download for Linux': '/downloads/infractl-linux-amd64',
  };
  const infractlLinks = Object.entries(infractlDownloads).map(([label, value]) => (
    <a key={value} href={value} download className="btn btn-base mr-2">
      {label}
    </a>
  ));

  return (
    <PageSection header="infractl (CLI)">
      <div className="mb-4 mx-2">{infractlLinks}</div>
      <div className="md:w-1/2 mx-2">
        <UserServiceAccountToken />
      </div>
    </PageSection>
  );
}
