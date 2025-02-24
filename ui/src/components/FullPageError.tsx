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
    <div className="pf-v5-u-display-flex pf-v5-u-align-items-center pf-v5-u-justify-content-center pf-v5-u-h-100">
      <IconComponent className="pf-v5-u-mr-md" size={64} />
      <span className="pf-v5-u-font-size-4xl">{message}</span>
    </div>
  );
}
