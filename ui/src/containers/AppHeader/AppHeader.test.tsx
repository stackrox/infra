import React from 'react';
import { render } from '@testing-library/react';
import { BrowserRouter as Router } from 'react-router-dom';

import AppHeader from './AppHeader';

test('renders app with the proper header', () => {
  const { getByText } = render(
    <Router>
      <AppHeader />
    </Router>,
  );
  const headerElement = getByText('Infra');
  expect(headerElement).toBeInTheDocument();
});
