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
    <div className="pf-u-display-flex pf-u-align-items-center pf-u-justify-content-center pf-u-h-100">
      <IconComponent className="pf-u-mr-md" size={64} />
      <span className="pf-u-font-size-4xl">{message}</span>
    </div>
  );
}
