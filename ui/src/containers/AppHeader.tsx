import React, { ReactElement } from 'react';

import ProductLogoTile from 'containers/ProductLogoTile';
import UserInfo from 'containers/UserInfo';

export default function AppHeader(): ReactElement {
  return (
    <header>
      <nav className="top-navigation flex justify-between bg-base-100">
        <div>
          <ProductLogoTile />
        </div>
        <div>
          <UserInfo />
        </div>
      </nav>
    </header>
  );
}
