import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';

import RHACSLogo from 'components/RHACSLogo';

export default function ProductLogoTile(): ReactElement {
  return (
    <Link to="/">
      <RHACSLogo />
    </Link>
  );
}
