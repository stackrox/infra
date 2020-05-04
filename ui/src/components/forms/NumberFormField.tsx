import React, { ReactElement } from 'react';
import { useField } from 'formik';
import NumericInput, { RcInputNumberProps } from 'rc-input-number';

import FormFieldError from './FormFieldError';

type Props = RcInputNumberProps & {
  name: string;
};

export default function NumberFormField(props: Props): ReactElement {
  const { name, ...inputNumberProps } = props;
  const [field, meta, helpers] = useField(name);

  return (
    <>
      <NumericInput
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        onChange={(v: number): void => helpers.setValue(v)}
        {...inputNumberProps} // eslint-disable-line react/jsx-props-no-spreading
      />
      <FormFieldError error={meta.error} touched={meta.touched} />
    </>
  );
}
