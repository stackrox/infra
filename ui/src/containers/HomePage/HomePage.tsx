import React, { ReactElement } from 'react';
import {
  Drawer,
  DrawerContent,
  DrawerContentBody,
  DrawerPanelContent,
  DrawerHead,
} from '@patternfly/react-core';

import LaunchPageSection from './LaunchPageSection';
import MyClustersPageSection from './MyClustersPageSection';

export default function HomePage(): ReactElement {
  return (
    <Drawer isExpanded isInline position="bottom" className="home-page">
      <DrawerContent
        panelContent={
          <DrawerPanelContent isResizable defaultSize="50%" minSize="150px">
            <DrawerHead>
              <MyClustersPageSection />
            </DrawerHead>
          </DrawerPanelContent>
        }
      >
        <DrawerContentBody>
          <LaunchPageSection />
        </DrawerContentBody>
      </DrawerContent>
    </Drawer>
  );
}
