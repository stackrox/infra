import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useMediaQuery } from 'react-responsive';

const defaultContextData = {
  isDarkMode: false,
  toggle: () => {},
};

export const ThemeContext = createContext(defaultContextData);
export const useTheme = () => useContext(ThemeContext);

const DARK_MODE_KEY = 'isDarkMode';

type ThemeState = {
  isDarkMode: boolean;
  hasThemeMounted: boolean;
};

// custom react hook to toggle dark mode across UI
function useEffectDarkMode(): [ThemeState, (next: ThemeState) => void] {
  const userPrefersDarkMode = useMediaQuery({ query: '(prefers-color-scheme: dark)' });
  const [themeState, setThemeState] = useState<ThemeState>({
    isDarkMode: userPrefersDarkMode,
    hasThemeMounted: false,
  });
  useEffect(() => {
    const darkModeValue = localStorage.getItem(DARK_MODE_KEY);
    let isDarkMode;
    // In the very beginning, default to using what the user prefers.
    if (darkModeValue === null) {
      isDarkMode = userPrefersDarkMode;
    } else {
      // It's always either 'true' or 'false', but if it's something unexpected,
      // default to light mode.
      isDarkMode = darkModeValue === 'true';
    }

    setThemeState({ isDarkMode, hasThemeMounted: true });
  }, [userPrefersDarkMode]);

  return [themeState, setThemeState];
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [themeState, setThemeState] = useEffectDarkMode();

  // to prevent theme flicker while getting theme from localStorage
  if (!themeState.hasThemeMounted) {
    return <div />;
  }

  if (themeState.isDarkMode) {
    document.documentElement.classList.add('pf-v6-theme-dark');
  } else {
    document.documentElement.classList.remove('pf-v6-theme-dark');
  }

  const toggle = () => {
    const darkModeToggled = !themeState.isDarkMode;

    localStorage.setItem(DARK_MODE_KEY, JSON.stringify(darkModeToggled));

    if (themeState.isDarkMode) {
      document.documentElement.classList.add('pf-v6-theme-dark');
    } else {
      document.documentElement.classList.remove('pf-v6-theme-dark');
    }

    setThemeState({ ...themeState, isDarkMode: darkModeToggled });
  };

  return (
    <ThemeContext.Provider
      value={{
        isDarkMode: themeState.isDarkMode,
        toggle,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
}
