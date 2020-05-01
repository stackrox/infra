import React, { ReactElement } from 'react';
import Tippy, { TippyProps } from '@tippy.js/react';

const defaultTippyTooltipProps = {
  arrow: true,
};

export default function Tooltip(props: TippyProps): ReactElement {
  // eslint-disable-next-line react/jsx-props-no-spreading
  return <Tippy {...defaultTippyTooltipProps} {...props} />;
}
