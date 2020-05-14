import React, { ReactElement, ReactNode } from 'react';

type Props = {
  header: ReactNode;
  className?: string;
  children: ReactNode;
};

export default function PageSection({ header, className = '', children }: Props): ReactElement {
  return (
    <div className="flex flex-col h-full min-h-0">
      <div className={`h-full overflow-auto ${className}`}>
        <h2 className="bg-base-0 border-b-2 border-base-400 capitalize font-600 mb-2 p-4 sticky text-4xl text-base-600 top-0">
          {header}
        </h2>
        <div className="flex flex-col p-4">{children}</div>
      </div>
    </div>
  );
}
