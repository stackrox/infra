import React, { ReactElement, ReactNode } from 'react';
import { Link } from 'react-router-dom';

type Props = {
  to: string;
  header: string;
  children: ReactNode;
  footer?: ReactNode;
  className?: string;
};

export default function LinkCard({
  to,
  header,
  children,
  footer,
  className = '',
}: Props): ReactElement {
  return (
    <Link
      className={`flex flex-col items-start h-32 w-64 p-2 border-2 border-base-400 shadow rounded font-600 bg-base-100 hover:bg-base-200 text-base-600 hover:text-base-700 ${className}`}
      to={to}
    >
      <h2 className="w-full pb-1 text-2xl border-b-2 border-base-500 mb-4">{header}</h2>
      {children}
      {!!footer && <div className="mt-auto">{footer}</div>}
    </Link>
  );
}
