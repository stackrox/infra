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
  min?: number;
  max?: number;
  disabled?: boolean;
};

export default function NumberFormField({
  name,
  id = `number-input-${name}`,
  required = false,
  label,
  min,
  max,
  disabled = false,
}: Props): ReactElement {
  const [field, meta, helpers] = useField(name);

  const input = (
    <input
      {...field} // eslint-disable-line react/jsx-props-no-spreading
      type="number"
      min={min}
      max={max}
      disabled={disabled}
      onChange={(e): void => {
        helpers.setValue(e.target.value || null); // force `null` instead of '' (otherwise it doesn't even trigger formik form validation)
      }}
      className={`bg-base-100 border-2 rounded p-2 border-base-300 font-600 text-base-600 leading-normal w-18 min-h-8 ${
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
