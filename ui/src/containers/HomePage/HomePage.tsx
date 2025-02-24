import React, { ReactElement } from 'react';

import { Divider } from '@patternfly/react-core';
import LaunchPageSection from './LaunchPageSection';
import MyClustersPageSection from './MyClustersPageSection';

export default function HomePage(): ReactElement {
  return (
    <>
      <LaunchPageSection />
      <Divider component="div" />
      <MyClustersPageSection />
    </>
  );
}
