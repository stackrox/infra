import React, { ReactElement, ReactNode } from 'react';
import {
  Masthead,
  MastheadMain,
  MastheadBrand,
  MastheadContent,
  ToolbarItem,
} from '@patternfly/react-core';

type Props = {
  logo: ReactNode;
  main: ReactNode;
  ending: ReactNode;
};

export default function AppHeaderLayout({ logo, main, ending }: Props): ReactElement {
  return (
    <Masthead>
      <MastheadMain>
        <MastheadBrand className="pf-u-mr-xl">{logo}</MastheadBrand>
        <ToolbarItem variant="separator" />
        {main}
      </MastheadMain>
      <MastheadContent className="pf-u-flex-direction-row-reverse">
        <div>{ending}</div>
      </MastheadContent>
    </Masthead>
  );
}
