import React, { ReactElement, ReactNode } from 'react';

type Props = {
  className?: string;
  children: ReactNode;
};

export default function TooltipOverlay({ className = '', children }: Props): ReactElement {
  return <div className={`rox-tooltip-overlay ${className}`}>{children}</div>;
}
