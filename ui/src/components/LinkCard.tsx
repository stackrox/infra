import React, { ReactElement, ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { Card, CardBody, CardFooter, CardHeader, CardTitle } from '@patternfly/react-core';

type Props = {
  to: string;
  header: ReactNode;
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
      <Card isClickable>
        <CardHeader selectableActions={{ to }}>
          <CardTitle>{header}</CardTitle>
        </CardHeader>
        <CardBody>{children}</CardBody>
        {!!footer && <CardFooter>{footer}</CardFooter>}
      </Card>
    </Link>
  );
}
