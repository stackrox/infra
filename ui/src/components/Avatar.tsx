import React, { ReactElement } from 'react';
import createInitials from 'initials';

type Props = {
  imageSrc?: string;
  name?: string;
  className?: string;
};

export default function Avatar({ imageSrc, name, className = '' }: Props): ReactElement {
  const finalClassName = `flex w-12 h-12 justify-center items-center rounded-full bg-secondary-500 ${className}`;
  const initials = name ? createInitials(name) : '--';

  return imageSrc ? (
    <img src={imageSrc} alt={initials} className={finalClassName} />
  ) : (
    <div className={finalClassName}>
      <span className="text-3xl font-800">{initials}</span>
    </div>
  );
}
