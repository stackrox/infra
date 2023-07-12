import React, { ReactElement } from 'react';
import { useField } from 'formik';
import { FormGroup, TextInput, ValidatedOptions, Popover } from '@patternfly/react-core';
import HelpIcon from '@patternfly/react-icons/dist/esm/icons/help-icon';

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

  const onChange = (_value: string, event: React.FormEvent<HTMLElement>) => {
    field.onChange(event);
  };

  return (
    <FormGroup
      label={label}
      fieldId={id}
      isRequired={required}
      labelIcon={
        helperText ? (
          <Popover bodyContent={<div>{helperText}</div>}>
            <button
              type="button"
              aria-label={`Help for ${name}`}
              onClick={(e) => e.preventDefault()}
              aria-describedby={id}
              className="pf-c-form__group-label-help"
            >
              <HelpIcon noVerticalAlign />
            </button>
          </Popover>
        ) : undefined
      }
      validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
      helperTextInvalid={meta.error}
      className="capitalize"
    >
      <TextInput
        id={id}
        name={name}
        onChange={onChange}
        type="text"
        value={field.value} // eslint-disable-line @typescript-eslint/no-unsafe-assignment
        placeholder={placeholder}
        isRequired={required}
        isDisabled={disabled}
        aria-describedby={`${id}-helper`}
        validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
        className={`bg-base-100 border-2 rounded p-2 my-2 border-base-300 font-600 text-base-600 leading-normal min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />
    </FormGroup>
  );
}
