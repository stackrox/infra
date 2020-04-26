import React, { ReactElement } from 'react';

import ProductLogoTile from 'components/ProductLogoTile';

export default function AppHeader(): ReactElement {
  return (
    <header>
      <nav className="top-navigation flex flex-1 justify-between relative bg-base-100">
        <div className="flex w-full">
          <ProductLogoTile />
        </div>
      </nav>
    </header>
  );
}
