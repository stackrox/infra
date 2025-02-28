import React, { ReactElement } from 'react';
import { Bullseye, EmptyState, Spinner } from '@patternfly/react-core';

export type FullPageSpinnerProps = {
  title?: string;
};

export default function FullPageSpinner({ title = 'Loading' }: FullPageSpinnerProps): ReactElement {
  return (
    <Bullseye>
      <EmptyState titleText={title} headingLevel="h4" icon={Spinner} />
    </Bullseye>
  );
}
