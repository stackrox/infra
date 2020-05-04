import React, { ReactElement, ReactNode } from 'react';

type Props = {
  logo: ReactNode;
  main: ReactNode;
  ending: ReactNode;
};

export default function AppHeaderLayout({ logo, main, ending }: Props): ReactElement {
  return (
    <header>
      <nav className="top-navigation flex bg-base-100">
        <div>{logo}</div>
        <div>{main}</div>
        <div className="ml-auto">{ending}</div>
      </nav>
    </header>
  );
}
