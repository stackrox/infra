import React, { ReactElement, ReactNode } from 'react';
import { Masthead, MastheadMain, MastheadBrand, MastheadContent } from '@patternfly/react-core';

type Props = {
  logo: ReactNode;
  main: ReactNode;
  ending: ReactNode;
};

export default function AppHeaderLayout({ logo, main, ending }: Props): ReactElement {
  return (
    <Masthead>
      <MastheadMain>
        <MastheadBrand>{logo}</MastheadBrand>
        {main}
      </MastheadMain>
      <MastheadContent className="pf-u-flex-direction-row-reverse">
        <div className="pf-u-float-right">{ending}</div>
      </MastheadContent>
    </Masthead>
    // <header>
    //   <nav className="top-navigation flex bg-base-100">
    //     <div>{logo}</div>
    //     <div>{main}</div>
    //     <div className="ml-auto">{ending}</div>
    //   </nav>
    // </header>
  );
}
