import React, { ReactElement } from 'react';
import { RingLoader } from 'react-spinners';

export default function FullPageSpinner(): ReactElement {
  return (
    <div className="pf-v5-u-display-flex pf-v5-u-flex-direction-column pf-v5-u-align-items-center pf-v5-u-justify-content-center pf-v5-u-h-100">
      <RingLoader loading size={128} color="currentColor" />
      <span className="pf-v5-u-font-size-4xl">Loading...</span>
    </div>
  );
}
