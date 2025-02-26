import React, { ReactElement } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Flex, Page } from '@patternfly/react-core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import UserAuthProvider from 'containers/UserAuthProvider';
import AppHeader from 'containers/AppHeader';
import HomePage from 'containers/HomePage';
import DownloadsPage from 'containers/DownloadsPage';
import LaunchClusterPage from 'containers/LaunchClusterPage';
import ClusterInfoPage from 'containers/ClusterInfoPage';
import FourOhFour from 'components/FourOhFour';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 10 * 1000,
    },
  },
});

function AppRoutes(): ReactElement {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/downloads" element={<DownloadsPage />} />
      <Route path="/launch/:flavorId" element={<LaunchClusterPage />} />
      <Route path="/cluster/:clusterId" element={<ClusterInfoPage />} />
      <Route path="*" element={<FourOhFour />} />
    </Routes>
  );
}

export default function App(): ReactElement {
  return (
    <Router>
      <UserAuthProvider>
        <QueryClientProvider client={queryClient}>
          <Flex
            direction={{ default: 'column' }}
            flexWrap={{ default: 'nowrap' }}
            className="pf-v6-u-h-100 pf-v6-u-w-100"
          >
            <Page masthead={<AppHeader />}>
              <AppRoutes />
            </Page>
          </Flex>
        </QueryClientProvider>
      </UserAuthProvider>
    </Router>
  );
}
