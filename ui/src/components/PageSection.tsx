import React, { ReactElement, ReactNode } from 'react';
import { Panel, PanelMain, PanelMainBody, PanelHeader, Divider } from '@patternfly/react-core';

type Props = {
  header: ReactNode;
  className?: string;
  children: ReactNode;
};

export default function PageSection({ header, className = '', children }: Props): ReactElement {
  return (
    <Panel className={className}>
      <PanelHeader className="pf-u-font-size-2xl">{header}</PanelHeader>
      <Divider />
      <PanelMain>
        <PanelMainBody>{children}</PanelMainBody>
      </PanelMain>
    </Panel>
  );
}
