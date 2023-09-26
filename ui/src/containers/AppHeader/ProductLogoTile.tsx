import React, { ReactElement } from 'react';
import { Flex, FlexItem } from '@patternfly/react-core';

import RHACSLogo from 'components/RHACSLogo';
import Version from './Version';

export default function ProductLogoTile(): ReactElement {
  return (
    <Flex alignItems={{ default: 'alignItemsCenter' }}>
      <FlexItem>
        <RHACSLogo />
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
