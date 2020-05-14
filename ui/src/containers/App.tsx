import React, { ReactElement } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import UserAuthProvider from 'containers/UserAuthProvider';
import { ThemeProvider } from 'containers/ThemeProvider';
import AppHeader from 'containers/AppHeader';
import HomePage from 'containers/HomePage';
import DownloadsPage from 'containers/DownloadsPage';
import LaunchClusterPage from 'containers/LaunchClusterPage';
import ClusterInfoPage from 'containers/ClusterInfoPage';
import FullPageError from 'components/FullPageError';

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
      <ThemeProvider>
        <UserAuthProvider>
          <div className="flex flex-col h-full bg-base-0">
            <AppHeader />
            <AppRoutes />
          </div>
        </UserAuthProvider>
      </ThemeProvider>
    </Router>
  );
}
