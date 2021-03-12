import React, { ReactElement, useState } from 'react';
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

export default function FileUploadFormField({
  name,
  id = `text-input-${name}`,
  required = false,
  label,
  placeholder = '',
  helperText = '',
  disabled = false,
}: Props): ReactElement {
  const [field, meta, helpers] = useField(name);
  const [reader, setReader] = useState<FileReader | null>(null);

  return (
    <div className="flex flex-col mb-4">
      <FormFieldLabel text={label} labelFor={id} required={required} />

      {helperText.length > 0 && <span className="font-400 text-base-600 my-1">{helperText}</span>}

      <input
        id={`${id}-filename`}
        type="file"
        name={`${name}-filename`}
        placeholder={placeholder}
        disabled={disabled}
        onChange={(e): void => {
          if (reader) {
            reader.abort();
          }
          if (e.target.files) {
            const thisReader = new FileReader();
            thisReader.addEventListener('loadstart', () => {
              helpers.setError('');
              helpers.setValue('');
            });
            thisReader.addEventListener('load', (loadEvent) => {
              if (loadEvent.target?.result) {
                helpers.setValue(loadEvent.target.result);
              }
            });
            thisReader.addEventListener('error', () => {
              helpers.setError('The file could not be read. Check permissions and try again.');
            });
            thisReader.readAsText(e.target.files[0]);
            setReader(thisReader);
          } else {
            setReader(null);
          }
        }}
        className={`bg-base-100 border-2 rounded p-2 my-2 border-base-300 font-600 text-base-600 leading-normal min-h-8 ${
          disabled ? 'bg-base-200' : 'hover:border-base-400'
        }`}
      />

      <input
        {...field} // eslint-disable-line react/jsx-props-no-spreading
        id={id}
        type="hidden"
        name={name}
      />

      <FormFieldError error={meta.error} touched />
    </div>
  );
}
