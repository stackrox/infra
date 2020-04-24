import React from 'react';
import { render } from '@testing-library/react';
import App from './App';

test('renders app with the proper header', () => {
  const { getByText } = render(<App />);
  const headerElement = getByText(/Setup Next UI/i);
  expect(headerElement).toBeInTheDocument();
});
