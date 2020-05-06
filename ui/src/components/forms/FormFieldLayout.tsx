import React, { ReactNode, ReactElement } from 'react';

type Props = {
  label: ReactNode;
  input: ReactNode;
  error?: ReactNode;
};

export default function FormFieldLayout({ label, input, error }: Props): ReactElement {
  return (
    <div className="flex flex-col mb-4">
      <div className="mb-1 text-base-600 font-700">{label}</div>
      <div className="w-full my-1">{input}</div>
      <div className="w-full my-1">{error}</div>
    </div>
  );
}
