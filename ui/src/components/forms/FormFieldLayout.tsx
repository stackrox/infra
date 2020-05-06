import React, { ReactNode, ReactElement } from 'react';

type Props = {
  label: ReactNode;
  input: ReactNode;
  error?: ReactNode;
};

export default function FormFieldLayout({ label, input, error }: Props): ReactElement {
  return (
    <div className="flex flex-col mb-4">
      <div className="w-full mb-2">{label}</div>
      <div className="w-full">{input}</div>
      <div className="w-full mt-2">{error}</div>
    </div>
  );
}
