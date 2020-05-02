import React, { ReactElement } from 'react';
import { Gift } from 'react-feather';

export default function HomePage(): ReactElement {
  return (
    <div className="flex flex-col flex-1 items-center justify-center">
      <Gift size={128} />
      <span className="text-6xl pt-10">Coming Soon</span>
    </div>
  );
}
