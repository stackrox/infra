import React, { ReactElement } from 'react';
import { RingLoader } from 'react-spinners';

export default function FullPageSpinner(): ReactElement {
  return (
    <div className="flex flex-col flex-1 w-full h-full items-center justify-center bg-base-0">
      <RingLoader loading size={128} color="currentColor" />
      <span className="text-6xl pt-10">Loading...</span>
    </div>
  );
}
