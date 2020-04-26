import React from 'react';

import StackRoxLogo from 'components/StackRoxLogo';
import Version from 'components/Version';

function App(): JSX.Element {
  return (
    <div>
      <header>
        <StackRoxLogo />
        <p>Here will go StackRox Setup Next UI!</p>
      </header>
      <Version />
    </div>
  );
}

export default App;
