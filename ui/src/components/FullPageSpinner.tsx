import React, { ReactElement } from 'react';
import { RingLoader } from 'react-spinners';

export default function FullPageSpinner(): ReactElement {
  return (
    <div className="pf-u-display-flex pf-u-flex-direction-column pf-u-align-items-center pf-u-justify-content-center pf-u-h-100">
      <RingLoader loading size={128} color="currentColor" />
      <span className="pf-u-font-size-4xl">Loading...</span>
    </div>
  );
}
