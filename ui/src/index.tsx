import React from 'react';
import { createRoot } from 'react-dom/client';

import './index.css';
import App from 'containers/App';

const container = document.getElementById('root') as HTMLElement;
const root = createRoot(container);

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
