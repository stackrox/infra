import React, { ReactElement } from 'react';
import { useField } from 'formik';

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

  return (
    <div className="flex flex-col mb-4">
      <FormFieldLabel text={label} labelFor={id} required={required} />

      <input
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        id={id}
        type="number"
        min={min}
        max={max}
        disabled={disabled}
        onChange={(e): void => {
          helpers.setValue(e.target.value || null); // force `null` instead of '' (otherwise it doesn't even trigger formik form validation)
        }}
        className={`bg-base-100 border-2 rounded p-2 my-2 border-base-300 font-600 text-base-600 leading-normal w-18 min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />

      <FormFieldError error={meta.error} touched={meta.touched} />
    </div>
  );
}
