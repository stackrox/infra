import React, { ReactElement } from 'react';
import { RingLoader } from 'react-spinners';

export default function FullPageSpinner(): ReactElement {
  return (
    <div className="pf-v6-u-display-flex pf-v6-u-flex-direction-column pf-v6-u-align-items-center pf-v6-u-justify-content-center pf-v6-u-h-100">
      <RingLoader loading size={128} color="currentColor" />
      <span className="pf-v6-u-font-size-4xl">Loading...</span>
    </div>
  );
}
