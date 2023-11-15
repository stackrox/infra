import React, { ReactElement, ReactNode } from 'react';
import { useField } from 'formik';
import { FormGroup, Popover, FileUpload, ValidatedOptions } from '@patternfly/react-core';
import { HelpIcon } from '@patternfly/react-icons';

type Props = {
  name: string;
  id?: string;
  required?: boolean;
  label: string;
  helper?: ReactNode;
};

export default function FileUploadFormField({
  name,
  id = `file-input-${name}`,
  required = false,
  label,
  helper,
}: Props): ReactElement {
  const [field, meta, helpers] = useField(name);
  const [filename, setFilename] = React.useState('');
  const [isLoading, setIsLoading] = React.useState(false);

  const handleFileInputChange = (
    event: React.ChangeEvent<HTMLInputElement> | React.DragEvent<HTMLElement>,
    file: File
  ) => {
    setFilename(file.name);
  };

  const handleTextOrDataChange = (value: string) => {
    helpers.setValue(value);
  };

  const handleClear = (_event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    setFilename('');
    helpers.setValue('');
  };

  const handleFileReadStarted = (_fileHandle: File) => {
    setIsLoading(true);
  };

  const handleFileReadFinished = (_fileHandle: File) => {
    setIsLoading(false);
  };

  return (
    <FormGroup
      label={label}
      fieldId={id}
      isRequired={required}
      labelIcon={
        helper ? (
          <Popover bodyContent={helper}>
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
    >
      <FileUpload
        id={id}
        value={field.value} // eslint-disable-line @typescript-eslint/no-unsafe-assignment
        filename={filename}
        filenamePlaceholder="Drag and drop a file or upload one"
        type="text"
        hideDefaultPreview
        onFileInputChange={handleFileInputChange}
        onDataChange={handleTextOrDataChange}
        onTextChange={handleTextOrDataChange}
        onReadStarted={handleFileReadStarted}
        onReadFinished={handleFileReadFinished}
        onClearClick={handleClear}
        isLoading={isLoading}
        browseButtonText="Upload"
        aria-describedby={`${id}-helper`}
        validated={meta.error ? ValidatedOptions.error : ValidatedOptions.default}
      />
    </FormGroup>
  );
}
