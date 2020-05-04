import React, { ReactNode, ReactElement } from 'react';

type Props = {
  label: ReactNode;
  children: ReactNode;
};

export default function Labeled({ label, children }: Props): ReactElement | null {
  if (!React.Children.count(children)) return null; // don't render w/o children
  return (
    <div className="mb-4">
      <div className="py-1 text-base-600 font-700">{label}</div>
      <div className="w-full py-1">{children}</div>
    </div>
  );
}
