import React, { ReactElement } from 'react';
import { Gift } from 'react-feather';

import { ThemeProvider } from 'components/ThemeProvider';
import AppHeader from 'components/AppHeader';

function App(): ReactElement {
  return (
    <ThemeProvider>
      <div className="flex flex-col h-full bg-base-0">
        <AppHeader />
        <div className="flex flex-col flex-1 items-center justify-center">
          <Gift size={128} />
          <span className="text-6xl pt-10">Coming Soon</span>
        </div>
      </div>
    </ThemeProvider>
  );
}

export default App;
