import React, { ReactElement } from 'react';

type Props = {
  text: string;
  /** will be used as htmlFor */
  labelFor: string;
  required?: boolean;
};

export default function FormFieldLabel({ text, labelFor, required = false }: Props): ReactElement {
  return (
    <label htmlFor={labelFor} className="text-base-600 font-700 capitalize">
      {text} {required && <i className="text-base-500 lowercase">(required)</i>}
    </label>
  );
}
