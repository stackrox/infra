import React, { ReactElement, ReactNode } from 'react';

type Props = {
  header: ReactNode;
  className?: string;
  children: ReactNode;
};

export default function PageSection({ header, className = '', children }: Props): ReactElement {
  return (
    <div className="border-t-3 border-base-400 flex flex-col h-full min-h-0">
      <div className={`h-full overflow-auto ${className}`}>
        <header className="border-b-2 border-base-400 px-4 py-2">
          <h2 className="bg-base-0 capitalize font-600 sticky text-4xl text-base-600 top-0">
            {header}
          </h2>
        </header>
        <div className="flex flex-col p-4">{children}</div>
      </div>
    </div>
  );
}
