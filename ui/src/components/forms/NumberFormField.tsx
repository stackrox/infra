import React, { ReactElement } from 'react';
import { useField } from 'formik';

import FormFieldError from './FormFieldError';

type Props = {
  name: string;
  min?: number;
  max?: number;
  disabled?: boolean;
};

export default function NumberFormField({ name, min, max, disabled = false }: Props): ReactElement {
  const [field, meta, helpers] = useField(name);

  return (
    <>
      <input
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        type="number"
        min={min}
        max={max}
        disabled={disabled}
        onChange={(e): void => {
          helpers.setValue(e.target.value || 0); // force 0 instead of ''
        }}
        className={`bg-base-100 border-2 rounded p-2 border-base-300 font-600 text-base-600 leading-normal w-18 min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />
      <FormFieldError error={meta.error} touched={meta.touched} />
    </>
  );
}
