import React, { ReactElement } from 'react';
import { AlertCircle } from 'react-feather';

type Props = {
  message: string;
};

export default function FullPageError({ message }: Props): ReactElement {
  return (
    <div className="flex flex-row w-full h-full items-center justify-center bg-base-0">
      <AlertCircle size={64} />
      <span className="pl-2 text-4xl">{message}</span>
    </div>
  );
}
