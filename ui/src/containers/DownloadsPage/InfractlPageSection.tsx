import React, { ReactElement } from 'react';

import PageSection from 'components/PageSection';
import UserServiceAccountToken from './UserServiceAccountToken';

export default function InfractlPageSection(): ReactElement {
  const infractlDownloads = {
    'Download for Mac': '/downloads/infractl-darwin-amd64',
    'Download for Linux': '/downloads/infractl-linux-amd64',
  };
  const infractlLinks = Object.entries(infractlDownloads).map(([label, value]) => (
    <a key={value} href={value} download={value} className="btn btn-base mr-2">
      {label}
    </a>
  ));

  return (
    <PageSection header="infractl (CLI)">
      <div className="mb-4">{infractlLinks}</div>
      <UserServiceAccountToken className="md:w-1/2" />
    </PageSection>
  );
}
