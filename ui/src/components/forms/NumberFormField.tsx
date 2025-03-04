import React, { ReactElement } from 'react';
import { useField } from 'formik';
import {
  FormGroup,
  FormHelperText,
  HelperText,
  HelperTextItem,
  NumberInput,
  ValidatedOptions,
} from '@patternfly/react-core';

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
    <FormGroup label={label} fieldId={id} isRequired={required} className="capitalize">
      <NumberInput
        id={id}
        name={name}
        onMinus={onMinus}
        onChange={onChange}
        onPlus={onPlus}
        type="text"
        value={typeof field.value === 'number' ? field.value : parseInt(field.value, 10)}
        min={min}
        max={max}
        isDisabled={disabled}
        validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
      />
      <FormHelperText>
        <HelperText>
          <HelperTextItem variant={meta.error ? ValidatedOptions.error : ValidatedOptions.default}>
            {meta.error}
          </HelperTextItem>
        </HelperText>
      </FormHelperText>
    </FormGroup>
  );
}
