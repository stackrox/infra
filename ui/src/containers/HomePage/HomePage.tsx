import React, { ReactElement } from 'react';

import LaunchPageSection from './LaunchPageSection';
import MyClustersPageSection from './MyClustersPageSection';

export default function HomePage(): ReactElement {
  return (
    <div className="overflow-y-scroll">
      <LaunchPageSection />
      <MyClustersPageSection />
    </div>
  );
}
