import React, { ReactElement, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardBody, CardFooter, CardHeader, CardTitle } from '@patternfly/react-core';

type Props = {
  to: string;
  header: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
};

export default function LinkCard({ to, header, children, footer }: Props): ReactElement {
  const navigate = useNavigate();
  return (
    <Card isClickable isCompact>
      <CardHeader
        selectableActions={{
          onClickAction: () => navigate(to),
          selectableActionAriaLabel: `Navigate to ${to}`,
        }}
      >
        <CardTitle>{header}</CardTitle>
      </CardHeader>
      <CardBody>{children}</CardBody>
      {!!footer && <CardFooter>{footer}</CardFooter>}
    </Card>
  );
}
