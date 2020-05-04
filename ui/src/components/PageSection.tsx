import React, { ReactElement, ReactNode } from 'react';

type Props = {
  header: string;
  className?: string;
  children: ReactNode;
};

export default function PageSection({ header, className = '', children }: Props): ReactElement {
  return (
    <div className={className}>
      <h1 className="pb-2 m-4 border-b-2 text-base-600 font-700 text-4xl capitalize">{header}</h1>
      <div className="ml-2 mr-2">{children}</div>
    </div>
  );
}
