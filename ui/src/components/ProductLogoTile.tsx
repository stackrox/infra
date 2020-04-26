import React, { ReactElement } from 'react';

import StackRoxLogo from 'components/shared/StackRoxLogo';
import Version from 'components/Version';

export default function ProductLogoTile(): ReactElement {
  return (
    <div className="flex flex-col items-center pb-1 px-4 border-r border-base-400">
      <div className="flex items-center">
        <StackRoxLogo />
        <div className="pl-1 pt-1 text-sm tracking-wide font-600 font-condensed uppercase">
          Infra
        </div>
      </div>
      <Version />
    </div>
  );
}
