import React, { ReactElement } from 'react';
import { EmptyState, EmptyStateBody } from '@patternfly/react-core';

type Props = {
  title?: string;
  message?: string;
};

export default function FullPageError({
  title = 'There was an unexpected error!',
  message = '',
}: Props): ReactElement {
  return (
    <EmptyState status="danger" variant="lg" titleText={title} headingLevel="h4">
      <EmptyStateBody>{message}</EmptyStateBody>
    </EmptyState>
  );
}
