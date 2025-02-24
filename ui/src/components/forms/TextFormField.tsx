import React, { ReactElement, ReactNode } from 'react';
import { useField } from 'formik';
import {
  FormGroup,
  FormHelperText,
  HelperText,
  HelperTextItem,
  Icon,
  Popover,
  TextInput,
  ValidatedOptions,
} from '@patternfly/react-core';
import { HelpIcon } from '@patternfly/react-icons';

type Props = {
  name: string;
  id?: string;
  required?: boolean;
  label: string;
  placeholder?: string;
  helper?: ReactNode;
  disabled?: boolean;
};

export default function TextFormField({
  name,
  id = `text-input-${name}`,
  required = false,
  label,
  placeholder = '',
  helper,
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
      labelHelp={
        helper ? (
          <Popover bodyContent={helper} className="form-help">
            <button
              type="button"
              aria-label={`Help for ${name}`}
              onClick={(e) => e.preventDefault()}
              aria-describedby={id}
              className="pf-v5-c-form__group-label-help"
            >
              <Icon>
                <HelpIcon />
              </Icon>
            </button>
          </Popover>
        ) : undefined
      }
    >
      <TextInput
        id={id}
        name={name}
        onChange={(event, _value: string) => onChange(_value, event)}
        type="text"
        value={field.value} // eslint-disable-line @typescript-eslint/no-unsafe-assignment
        placeholder={placeholder}
        isRequired={required}
        isDisabled={disabled}
        aria-describedby={`${id}-helper`}
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
