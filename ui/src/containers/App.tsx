import React, { ReactElement } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import UserAuthProvider from 'containers/UserAuthProvider';
import { ThemeProvider } from 'containers/ThemeProvider';
import AppHeader from 'containers/AppHeader';
import HomePage from 'containers/HomePage';

function App(): ReactElement {
  return (
    <Router>
      <ThemeProvider>
        <UserAuthProvider>
          <div className="flex flex-col h-full bg-base-0">
            <AppHeader />
            <Routes>
              <Route path="/" element={<HomePage />} />
            </Routes>
          </div>
        </UserAuthProvider>
      </ThemeProvider>
    </Router>
  );
}

export default App;
