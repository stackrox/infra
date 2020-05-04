import React, { ReactElement } from 'react';
import { useField } from 'formik';

import FormFieldError from './FormFieldError';

type Props = {
  name: string;
  placeholder?: string;
  disabled?: boolean;
};

export default function TextFormField({
  name,
  placeholder = '',
  disabled = false,
}: Props): ReactElement {
  const [field, meta] = useField(name);

  return (
    <>
      <input
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        type="text"
        name={name}
        placeholder={placeholder}
        disabled={disabled}
        className={`bg-base-100 border-2 rounded p-2 border-base-300 w-full font-600 text-base-600 leading-normal min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />
      <FormFieldError error={meta.error} touched={meta.touched} />
    </>
  );
}
