import React, { ReactElement } from 'react';
import { useField } from 'formik';

import FormFieldLabel from './FormFieldLabel';
import FormFieldError from './FormFieldError';

type Props = {
  name: string;
  id?: string;
  required?: boolean;
  label: string;
  placeholder?: string;
  helperText?: string;
  disabled?: boolean;
};

export default function TextFormField({
  name,
  id = `text-input-${name}`,
  required = false,
  label,
  placeholder = '',
  helperText = '',
  disabled = false,
}: Props): ReactElement {
  const [field, meta] = useField(name);

  return (
    <div className="flex flex-col mb-4">
      <FormFieldLabel text={label} labelFor={id} required={required} />

      {helperText.length > 0 && <span className="font-400 text-base-600 my-1">{helperText}</span>}

      <input
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        id={id}
        type="text"
        name={name}
        placeholder={placeholder}
        disabled={disabled}
        className={`bg-base-100 border-2 rounded p-2 my-2 border-base-300 font-600 text-base-600 leading-normal min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />

      <FormFieldError error={meta.error} touched={meta.touched} />
    </div>
  );
}
