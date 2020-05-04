import React, { ReactElement, ReactNode } from 'react';

type Props = {
  header: ReactNode | string;
  className?: string;
  children: ReactNode;
};

export default function PageSection({ header, className, children }: Props): ReactElement {
  const renderedHeader = (
    <div className="pb-2 m-4 border-b-2">
      {typeof header === 'string' ? (
        <h1 className="text-base-600 font-700 text-4xl">{header}</h1>
      ) : (
        header
      )}
    </div>
  );

  return (
    <div className={`${className || ''}`}>
      {renderedHeader}
      <div className="ml-2 mr-2">{children}</div>
    </div>
  );
}
