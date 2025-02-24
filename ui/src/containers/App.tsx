import React, { ReactElement } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import UserAuthProvider from 'containers/UserAuthProvider';
import AppHeader from 'containers/AppHeader';
import HomePage from 'containers/HomePage';
import DownloadsPage from 'containers/DownloadsPage';
import LaunchClusterPage from 'containers/LaunchClusterPage';
import ClusterInfoPage from 'containers/ClusterInfoPage';
import FullPageError from 'components/FullPageError';
import { Flex } from '@patternfly/react-core';

function AppRoutes(): ReactElement {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/downloads" element={<DownloadsPage />} />
      <Route path="/launch/:flavorId" element={<LaunchClusterPage />} />
      <Route path="/cluster/:clusterId" element={<ClusterInfoPage />} />
      <Route path="*" element={<FullPageError message="This page doesn't seem to exist." />} />
    </Routes>
  );
}

export default function App(): ReactElement {
  return (
    <Router>
      <UserAuthProvider>
        <Flex
          direction={{ default: 'column' }}
          flexWrap={{ default: 'nowrap' }}
          className="pf-v6-u-h-100 pf-v6-u-w-100"
        >
          <AppHeader />
          <AppRoutes />
        </Flex>
      </UserAuthProvider>
    </Router>
  );
}
