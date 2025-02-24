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
    <div className="pf-v6-u-display-flex pf-v6-u-align-items-center pf-v6-u-justify-content-center pf-v6-u-h-100">
      <IconComponent className="pf-v6-u-mr-md" size={64} />
      <span className="pf-v6-u-font-size-4xl">{message}</span>
    </div>
  );
}
