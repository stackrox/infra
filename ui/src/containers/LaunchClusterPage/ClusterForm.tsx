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
import Labeled from 'components/Labeled';
import FormFieldLabel from 'components/FormFieldLabel';
import TextFormField from 'components/forms/TextFormField';
import NumberFormField from 'components/forms/NumberFormField';

const clusterService = new ClusterServiceApi(configuration);

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const schemasByParameterName: { [key: string]: yup.Schema<any> } = {
  name: yup
    .string()
    .default('')
    .required('Required')
    .min(3, 'Too short')
    .max(40, 'Too long')
    .matches(
      /^(?:[a-z](?:[-a-z0-9]{0,38}[a-z0-9])?)$/, // this is what GKE expects
      "Only alphanumerics and '-' allowed, must start with a letter and end with an alphanumeric"
    ),
  nodes: yup
    .number()
    .default(2)
    .required('Required')
    .min(1, 'Must be at least 1')
    .max(10, 'Be cost effective, please'),
};

type FlavorParameters = { [key: string]: V1Parameter };

function createParameterSchemas(parameters: FlavorParameters): object {
  return Object.keys(parameters).reduce<object>((fields, param) => {
    if (!schemasByParameterName[param]) throw new Error(`Unknown parameter type "${param}"`);
    return {
      ...fields,
      [param]: schemasByParameterName[param],
    };
  }, {});
}

function createInitialParameterValues(parameters: FlavorParameters): object {
  return Object.keys(parameters).reduce<object>((fields, param) => {
    if (!schemasByParameterName[param]) throw new Error(`Unknown parameter type "${param}"`);
    return {
      ...fields,
      [param]: schemasByParameterName[param].default(),
    };
  }, {});
}

// backend expects every parameter value to be a string, i.e. instead of 3 to be "3"
function adjustParametersBeforeSubmit(parameterValues: FormikValues): { [key: string]: string } {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return mapValues(parameterValues, (value: any) => `${value}`);
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

function ParameterFormField(props: { parameter: V1Parameter }): ReactElement {
  const { parameter } = props;
  const schema = parameter.Name && schemasByParameterName[parameter.Name];
  if (!parameter.Name || !schema) throw new Error(`Unknown parameter "${parameter.Name}"`);

  switch (schema.type) {
    case 'string':
      return <TextFormField name={`Parameters.${parameter.Name}`} />;
    case 'number':
      return (
        <NumberFormField
          name={`Parameters.${parameter.Name}`}
          min={getSchemaTestParamValue(schema, 'min')}
          max={getSchemaTestParamValue(schema, 'max')}
        />
      );
    default:
      throw new Error(`Unknown type "${schema.type}" for parameter "${parameter.Name}"`);
  }
}

function FormContent(props: { flavorParameters: FlavorParameters }): ReactElement {
  const { isSubmitting } = useFormikContext();
  const { flavorParameters } = props;
  const parameterFields = Object.entries(flavorParameters).map(([param, metadata]) => (
    <Labeled
      key={param}
      label={
        <FormFieldLabel text={metadata.Description || param} required className="capitalize" />
      }
    >
      <ParameterFormField parameter={metadata} />
    </Labeled>
  ));
  return (
    <>
      {parameterFields}

      <Labeled label={<FormFieldLabel text="Description" />}>
        <TextFormField name="Description" />
      </Labeled>

      <button type="submit" className="btn btn-base" disabled={isSubmitting}>
        {isSubmitting ? <ClipLoader size={16} color="currentColor" /> : 'Launch'}
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
  const schema = yup.object().shape({
    ID: yup.string().required(),
    Description: yup.string().default(''),
    Parameters: yup.object().shape(createParameterSchemas(flavorParameters)),
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
        <FormContent flavorParameters={flavorParameters} />
      </Form>
    </Formik>
  );
}
