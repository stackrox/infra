import React, { ReactElement } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Tool } from 'react-feather';

import UserAuthProvider from 'containers/UserAuthProvider';
import { ThemeProvider } from 'containers/ThemeProvider';
import AppHeader from 'containers/AppHeader';
import HomePage from 'containers/HomePage';
import DownloadsPage from 'containers/DownloadsPage';
import LaunchClusterPage from 'containers/LaunchClusterPage';
import FullPageError from 'components/FullPageError';

function AppRoutes(): ReactElement {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/downloads" element={<DownloadsPage />} />
      <Route path="/launch/:flavorId" element={<LaunchClusterPage />} />
      <Route
        path="/cluster/:clusterId"
        element={
          <FullPageError
            message="Cluster page is under construction. Use infractl instead."
            IconComponent={Tool}
          />
        }
      />
      <Route
        path="*"
        element={<FullPageError message="WIP. Pardon our dust." IconComponent={Tool} />}
      />
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
