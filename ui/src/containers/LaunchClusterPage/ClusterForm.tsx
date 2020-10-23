/**
 * This file is mostly an experiment in generating dynamic forms based on the data provided by the
 * server. For now API doesn't return enough metadata about flavor parameters, so some assumptions
 * and hard-coding has to be made. Because of its experimental nature, both components and helper
 * functions (e.g to work with Yup schemas) are kept confined to this file.
 *
 * Likely most of this file will be rewritten once API supports flavor parameters metadata.
 */

import React, { useState, ReactElement } from 'react';
import { Formik, Form, FormikValues, FormikHelpers, useFormikContext } from 'formik';
import * as yup from 'yup';
import { mapValues } from 'lodash';
import { ClipLoader } from 'react-spinners';

import { ClusterServiceApi, V1Parameter } from 'generated/client';
import configuration from 'client/configuration';
import TextFormField from 'components/forms/TextFormField';
import NumberFormField from 'components/forms/NumberFormField';
import { UploadCloud } from 'react-feather';

const clusterService = new ClusterServiceApi(configuration);

function helpByParameterName(name?: string): string {
  const help: { [key: string]: string } = {
    name:
      "Only lowercase letters, numbers, and '-' allowed, must start with a letter and end with a letter or number",
  };

  if (name && name in help) {
    return help[name];
  }
  return '';
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const schemasByParameterName: { [key: string]: yup.Schema<any> } = {
  name: yup
    .string()
    .min(3, 'Too short')
    .max(40, 'Too long')
    .matches(
      /^(?:[a-z](?:[-a-z0-9]{0,38}[a-z0-9])?)$/, // this is what GKE expects
      'The input value does not match the criteria above. Please correct this form field.'
    ),
  nodes: yup
    .number()
    .transform((v) => (Number.isNaN(v) ? 0.1 : (v as number))) // workaround https://github.com/jquense/yup/issues/66
    .integer('Must be an integer')
    .min(1, 'Must be at least 1')
    .max(50, 'Be cost effective, please'),
};

type FlavorParameters = { [key: string]: V1Parameter };
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ParameterSchemas = { [key: string]: yup.Schema<any> };

function createParameterSchemas(parameters: FlavorParameters): ParameterSchemas {
  return Object.keys(parameters).reduce<object>((fields, param) => {
    let thisParamSchema;
    if (schemasByParameterName[param]) {
      thisParamSchema = schemasByParameterName[param];
    } else {
      thisParamSchema = yup.string();
    }
    if (!parameters[param].Optional) {
      thisParamSchema = (thisParamSchema as yup.MixedSchema).required('Required');
    }
    return {
      ...fields,
      [param]: thisParamSchema,
    };
  }, {}) as ParameterSchemas;
}

function createInitialParameterValues(parameters: FlavorParameters): object {
  return Object.keys(parameters).reduce<object>((fields, param) => {
    return {
      ...fields,
      [param]: parameters[param].Optional && parameters[param].Value ? parameters[param].Value : '',
    };
  }, {});
}

// backend expects every parameter value to be a string, i.e. instead of 3 to be "3"
function adjustParametersBeforeSubmit(parameterValues: FormikValues): { [key: string]: string } {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return mapValues(parameterValues, (value: any) => `${value}`.trim());
}

/**
 * Gets test param value from the schema description by test name and
 * an optional param name for this test.
 *
 * @template T type the schema describes
 * @template V test param value type
 * @param {yup.Schema<T>} schema Yup schema
 * @param {string} testName name of the test in the schema
 * @param {string} [testParamName=testName] name of the test parameter to return value of
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getSchemaTestParamValue<T = any, V = any>(
  schema: yup.Schema<T>,
  testName: string,
  testParamName: string = testName
): V {
  const { tests } = schema.describe();
  const test = tests.find((t) => t.name === testName);
  return test?.params[testParamName] as V;
}

function getFormLabelFromParameter(parameter: V1Parameter): string {
  return parameter.Description || parameter.Name || '';
}

function ParameterFormField(props: {
  parameter: V1Parameter;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  schema: yup.Schema<any>;
}): ReactElement {
  const { parameter, schema } = props;

  const required = !parameter.Optional;

  switch (schema.type) {
    case 'string':
      return (
        <TextFormField
          name={`Parameters.${parameter.Name}`}
          label={getFormLabelFromParameter(parameter)}
          helperText={parameter.Help || helpByParameterName(parameter.Name)}
          required={required}
        />
      );
    case 'number':
      return (
        <NumberFormField
          name={`Parameters.${parameter.Name}`}
          label={getFormLabelFromParameter(parameter)}
          required={required}
          min={getSchemaTestParamValue(schema, 'min')}
          max={getSchemaTestParamValue(schema, 'max')}
        />
      );
    default:
      throw new Error(`Unknown type "${schema.type}" for parameter "${parameter.Name}"`);
  }
}

function getOrderFromParameter(parameter: V1Parameter): number {
  return parameter.Order || 0;
}

function getSchemaForParameter(
  parameterSchemas: ParameterSchemas,
  parameter: V1Parameter
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): yup.Schema<any> {
  if (parameter.Name && parameter.Name in parameterSchemas) {
    return parameterSchemas[parameter.Name];
  }
  return yup.string();
}

function FormContent(props: {
  flavorParameters: FlavorParameters;
  parameterSchemas: ParameterSchemas;
}): ReactElement {
  const { isSubmitting } = useFormikContext();
  const { flavorParameters, parameterSchemas } = props;
  const parameterFields = Object.values(flavorParameters)
    .sort((a: V1Parameter, b: V1Parameter) => getOrderFromParameter(a) - getOrderFromParameter(b))
    .map((parameter) => (
      <ParameterFormField
        key={parameter.Name}
        parameter={parameter}
        schema={getSchemaForParameter(parameterSchemas, parameter)}
      />
    ));

  const launchBtnContent = (
    <>
      <UploadCloud size={16} className="mr-2" />
      Launch
    </>
  );

  return (
    <>
      {parameterFields}

      <TextFormField name="Description" label="Description" />

      <button type="submit" className="btn btn-base" disabled={isSubmitting}>
        {isSubmitting ? <ClipLoader size={16} color="currentColor" /> : launchBtnContent}
      </button>
    </>
  );
}

type Props = {
  flavorId: string;
  flavorParameters: FlavorParameters;
  onClusterCreated: (clusterId: string) => void;
};

export default function ClusterForm({
  flavorId,
  flavorParameters,
  onClusterCreated,
}: Props): ReactElement {
  const parameterSchemas = createParameterSchemas(flavorParameters);
  const schema = yup.object().shape({
    ID: yup.string().required(),
    Description: yup.string().default(''),
    Parameters: yup.object().shape<object>(parameterSchemas),
  });
  const initialValues: FormikValues = {
    ID: flavorId,
    Description: '',
    Parameters: createInitialParameterValues(flavorParameters),
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [error, setError] = useState<any>();

  const onSubmit = async (
    values: FormikValues,
    actions: FormikHelpers<FormikValues>
  ): Promise<void> => {
    try {
      const response = await clusterService.create({
        ...values,
        Parameters: adjustParametersBeforeSubmit(values.Parameters),
      });

      const { id } = response.data;
      if (!id) throw new Error('Server returned empty cluster ID');
      onClusterCreated(id);
    } catch (e) {
      setError(e);
    } finally {
      actions.setSubmitting(false);
    }
  };

  return (
    <Formik initialValues={initialValues} validationSchema={schema} onSubmit={onSubmit}>
      <Form className="md:w-1/3">
        {error && (
          <div className="p-2 mb-2 bg-alert-200">
            {`[Server Error] ${error.message || 'Cluster creation request failed'}`}
            {error.response?.data?.error && ` (${error.response.data.error})`}
          </div>
        )}
        <FormContent flavorParameters={flavorParameters} parameterSchemas={parameterSchemas} />
      </Form>
    </Formik>
  );
}
