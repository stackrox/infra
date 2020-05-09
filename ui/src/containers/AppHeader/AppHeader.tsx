import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';
import { Terminal } from 'react-feather';

import AppHeaderLayout from 'components/AppHeaderLayout';
import ProductLogoTile from './ProductLogoTile';
import UserInfo from './UserInfo';

export default function AppHeader(): ReactElement {
  const mainArea = (
    <div className="flex h-full items-center ml-4">
      <Link to="/downloads" className="btn btn-base">
        <Terminal size={16} className="mr-2" />
        infractl
      </Link>
    </div>
  );

  return <AppHeaderLayout logo={<ProductLogoTile />} main={mainArea} ending={<UserInfo />} />;
}
