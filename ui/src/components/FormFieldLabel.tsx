import React, { ReactElement } from 'react';

type Props = {
  text: string;
  required?: boolean;
  className?: string;
};

export default function FormFieldLabel({
  text,
  required = false,
  className = '',
}: Props): ReactElement {
  return (
    <p className={className}>
      {text} {required && <i className="text-base-500 lowercase">(required)</i>}
    </p>
  );
}
