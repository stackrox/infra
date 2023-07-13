import React, { ReactElement } from 'react';
import { useField } from 'formik';
import { FormGroup, NumberInput, ValidatedOptions } from '@patternfly/react-core';

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

  const onMinus = () => {
    const newValue = (field.value || 0) - 1;
    helpers.setValue(newValue);
  };

  const onChange = (event: React.FormEvent<HTMLInputElement>) => {
    const { value } = event.target as HTMLInputElement;
    helpers.setValue(value === '' ? value : +value);
  };

  const onPlus = () => {
    const newValue = (+(field.value as number) || 0) + 1;
    helpers.setValue(newValue);
  };

  return (
    <FormGroup
      label={label}
      fieldId={id}
      isRequired={required}
      validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
      helperTextInvalid={meta.error}
      className="capitalize"
    >
      <NumberInput
        id={id}
        name={name}
        onMinus={onMinus}
        onChange={onChange}
        onPlus={onPlus}
        type="text"
        value={field.value} // eslint-disable-line @typescript-eslint/no-unsafe-assignment
        min={min}
        max={max}
        isDisabled={disabled}
        validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
        className={`bg-base-100 border-2 rounded p-2 my-2 border-base-300 font-600 text-base-600 leading-normal min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />
    </FormGroup>
  );
}
