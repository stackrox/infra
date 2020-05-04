import React, { ReactElement } from 'react';
import { Gift } from 'react-feather';

import UserAuthProvider from 'containers/UserAuthProvider';
import { ThemeProvider } from 'containers/ThemeProvider';
import AppHeader from 'containers/AppHeader';

function App(): ReactElement {
  return (
    <ThemeProvider>
      <UserAuthProvider>
        <div className="flex flex-col h-full bg-base-0">
          <AppHeader />
          <div className="flex flex-col flex-1 items-center justify-center">
            <Gift size={128} />
            <span className="text-6xl pt-10">Coming Soon</span>
          </div>
        </div>
      </UserAuthProvider>
    </ThemeProvider>
  );
}

export default App;
