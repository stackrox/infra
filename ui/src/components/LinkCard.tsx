import React, { ReactElement, ReactNode } from 'react';
import { Link } from 'react-router-dom';
import {
  Panel,
  PanelMain,
  PanelMainBody,
  PanelHeader,
  Divider,
  PanelFooter,
} from '@patternfly/react-core';

type Props = {
  to: string;
  header: string;
  children: ReactNode;
  footer?: ReactNode;
  className?: string;
};

export default function LinkCard({
  to,
  header,
  children,
  footer,
  className = '',
}: Props): ReactElement {
  return (
    <Link className={className} to={to}>
      <Panel isScrollable variant="raised">
        <PanelHeader>{header}</PanelHeader>
        <Divider />
        <PanelMain tabIndex={0}>
          <PanelMainBody>{children} </PanelMainBody>
        </PanelMain>
        {!!footer && <PanelFooter>{footer}</PanelFooter>}
      </Panel>
    </Link>
  );
}
