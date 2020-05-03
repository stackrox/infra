import React, { ReactElement } from 'react';
import { AlertCircle, Icon } from 'react-feather';

type Props = {
  message: string;
  IconComponent?: Icon;
};

export default function FullPageError({
  message,
  IconComponent = AlertCircle,
}: Props): ReactElement {
  return (
    <div className="flex flex-row w-full h-full items-center justify-center bg-base-0">
      <IconComponent size={64} />
      <span className="pl-2 text-4xl">{message}</span>
    </div>
  );
}
