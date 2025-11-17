import React, { ReactElement } from 'react';
import { Button, Tooltip } from '@patternfly/react-core';
import { useTheme } from 'utils/ThemeProvider';
import { MoonIcon, SunIcon } from '@patternfly/react-icons';

const ThemeToggleButton = (): ReactElement => {
  const themeState = useTheme();
  const tooltipText = themeState.isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode';

  return (
    <Tooltip content={<div>{tooltipText}</div>} position="bottom">
      <Button aria-label="Invert theme" onClick={themeState.toggle} variant="control">
        {themeState.isDarkMode ? <SunIcon /> : <MoonIcon />}
      </Button>
    </Tooltip>
  );
};

export default ThemeToggleButton;
