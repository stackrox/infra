import React, { ReactElement } from 'react';

type Props = {
  error?: string;
  touched: boolean;
};

export default function FormFieldError({ error, touched }: Props): ReactElement | null {
  return error && touched ? <div className="text-alert-400">{error}</div> : null;
}
