import React, { ReactElement } from 'react';

import LaunchPageSection from './LaunchPageSection';
import MyClustersPageSection from './MyClustersPageSection';

export default function HomePage(): ReactElement {
  return (
    <>
      <LaunchPageSection />
      <MyClustersPageSection />
    </>
  );
}
