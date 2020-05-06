import React, { ReactElement } from 'react';
import { useField } from 'formik';

import FormFieldLayout from './FormFieldLayout';
import FormFieldLabel from './FormFieldLabel';
import FormFieldError from './FormFieldError';

type Props = {
  name: string;
  id?: string;
  required?: boolean;
  label: string;
  placeholder?: string;
  disabled?: boolean;
};

export default function TextFormField({
  name,
  id = `text-input-${name}`,
  required = false,
  label,
  placeholder = '',
  disabled = false,
}: Props): ReactElement {
  const [field, meta] = useField(name);

  const input = (
    <input
      {...field} // eslint-disable-line react/jsx-props-no-spreading
      id={id}
      type="text"
      name={name}
      placeholder={placeholder}
      disabled={disabled}
      className={`bg-base-100 border-2 rounded p-2 border-base-300 w-full font-600 text-base-600 leading-normal min-h-8 ${
        disabled ? 'bg-base-200' : 'hover:border-base-400'
      }`}
    />
  );

  return (
    <FormFieldLayout
      label={<FormFieldLabel text={label} labelFor={id} required={required} />}
      input={input}
      error={<FormFieldError error={meta.error} touched={meta.touched} />}
    />
  );
}
