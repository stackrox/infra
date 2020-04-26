import React, { createContext, useContext, ReactElement, ReactNode } from 'react';

export enum Theme {
  Dark = 0,
  Light = 1,
}

export interface ThemeContextData {
  theme: Theme;
}

const defaultContextData: ThemeContextData = {
  theme: Theme.Dark,
};

const ThemeContext = createContext(defaultContextData);

/**
 * Returns the theme from the React context.
 * *Note: currently only dark theme is supported.*
 */
const useTheme = (): ThemeContextData => useContext(ThemeContext);

type Props = {
  children: ReactNode;
};

/**
 * React context based theme provider.
 */
function ThemeProvider({ children }: Props): ReactElement {
  return <ThemeContext.Provider value={defaultContextData}>{children}</ThemeContext.Provider>;
}

export { ThemeProvider, useTheme };
