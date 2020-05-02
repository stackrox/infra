import React from 'react';
import { render } from '@testing-library/react';
import AppHeader from './AppHeader';

test('renders app with the proper header', () => {
  const { getByText } = render(<AppHeader />);
  const headerElement = getByText(/Infra/i);
  expect(headerElement).toBeInTheDocument();
});
