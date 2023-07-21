import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';
import { Flex, FlexItem } from '@patternfly/react-core';

import RHACSLogo from 'components/RHACSLogo';
import Version from './Version';

export default function ProductLogoTile(): ReactElement {
  return (
    <Flex className="pf-u-align-items-center">
      <FlexItem>
        <Link to="/">
          <RHACSLogo />
        </Link>
      </FlexItem>
      <FlexItem>
        <span className="pf-u-font-family-redhatVF-heading-sans-serif pf-u-font-size-lg">
          Infra
        </span>
        <Version />
      </FlexItem>
    </Flex>
  );
}
