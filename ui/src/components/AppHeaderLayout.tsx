import React, { ReactElement, ReactNode } from 'react';
import {
  Masthead,
  MastheadMain,
  MastheadLogo,
  MastheadBrand,
  MastheadContent,
  ToolbarItem,
  Flex,
} from '@patternfly/react-core';
import Version from 'containers/AppHeader/Version';

type Props = {
  logo: ReactNode;
  main: ReactNode;
  ending: ReactNode;
};

export default function AppHeaderLayout({ logo, main, ending }: Props): ReactElement {
  return (
    <Masthead>
      <MastheadMain>
        <MastheadBrand>
          <MastheadLogo component="a" className="pf-v6-u-mr-xl">
            {logo}
          </MastheadLogo>
        </MastheadBrand>
      </MastheadMain>
      <MastheadContent>
        <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsNone' }}>
          <span className="pf-v6-u-font-family-heading pf-v6-u-font-size-lg">Infra</span>
          <Version />
        </Flex>
        <ToolbarItem variant="separator" />
        <Flex
          className="pf-v6-u-flex-grow-1"
          alignItems={{ default: 'alignItemsCenter' }}
          justifyContent={{ default: 'justifyContentSpaceBetween' }}
        >
          {main}
          <div>{ending}</div>
        </Flex>
      </MastheadContent>
    </Masthead>
  );
}
