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
import { UploadCloud } from 'react-feather';

import { ClusterServiceApi, V1Parameter } from 'generated/client';
import configuration from 'client/configuration';
import FileUploadFormField from 'components/forms/FileUploadFormField';
import TextFormField from 'components/forms/TextFormField';
import NumberFormField from 'components/forms/NumberFormField';
import { useUserAuth } from 'containers/UserAuthProvider';
import assertDefined from 'utils/assertDefined';
import { generateClusterName } from 'utils/cluster.utils';

const clusterService = new ClusterServiceApi(configuration);

function helpByParameterName(name?: string): string {
  const help: { [key: string]: string } = {
    name:
      "You can use the generated name, or type in your own. Only lowercase letters, numbers, and '-' allowed, must start with a letter and end with a letter or number. The name must be between 3 and 28 characters long.",
  };

  if (name && name in help) {
    return help[name];
  }
  return '';
}

const schemasByParameterName: { [key: string]: yup.BaseSchema } = {
  // name length is restricted by certificate generation
  name: yup
    .string()
    .min(3, 'Too short')
    .max(28, 'Too long')
    .matches(
      /^(?:[a-z](?:[a-z0-9-]{1,26}[a-z0-9]))$/,
      'The input value does not match its requirements. See the (?) section for details.'
    ),
  nodes: yup
    .number()
    .transform((v: unknown) => (Number.isNaN(v) ? 0.1 : v)) // workaround https://github.com/jquense/yup/issues/66
    .integer('Must be an integer')
    .min(1, 'Must be at least 1')
    .max(50, 'Be cost effective, please'),
};

type FlavorParameters = { [key: string]: V1Parameter };
type ParameterSchemas = { [key: string]: yup.BaseSchema };

function createParameterSchemas(parameters: FlavorParameters): ParameterSchemas {
  return Object.keys(parameters).reduce<Record<string, unknown>>((fields, param) => {
    let thisParamSchema: yup.BaseSchema;
    if (schemasByParameterName[param]) {
      thisParamSchema = schemasByParameterName[param];
    } else {
      thisParamSchema = yup.string();
    }
    if (!parameters[param].Optional) {
      thisParamSchema = thisParamSchema.required('Required') as yup.BaseSchema;
    }
    return {
      ...fields,
      [param]: thisParamSchema,
    };
  }, {}) as ParameterSchemas;
}

function createInitialParameterValues(parameters: FlavorParameters): Record<string, unknown> {
  return Object.keys(parameters).reduce<Record<string, unknown>>(
    (fields, param) => ({
      ...fields,
      [param]: parameters[param].Value ? parameters[param].Value : '',
    }),
    {}
  );
}

// backend expects every parameter value to be a string, i.e. instead of 3 to be "3"
function adjustParametersBeforeSubmit(parameterValues: FormikValues): { [key: string]: string } {
  return mapValues(parameterValues, (value: unknown) => String(value).trim());
}

/**
 * Gets test param value from the schema description by test name and
 * an optional param name for this test.
 *
 * @template T type the schema describes
 * @template V test param value type
 * @param {yup.BaseSchema<T>} schema Yup schema
 * @param {string} testName name of the test in the schema
 * @param {string} [testParamName=testName] name of the test parameter to return value of
 */
function getSchemaTestParamValue<T = unknown, V = unknown>(
  schema: yup.BaseSchema<T>,
  testName: string,
  testParamName: string = testName
): V | undefined {
  const { tests } = schema.describe();
  const test = tests.find((t) => t.name === testName);
  return test?.params && (test?.params[testParamName] as V);
}

function getFormLabelFromParameter(parameter: V1Parameter): string {
  return parameter.Description || parameter.Name || '';
}

function ParameterFormField(props: {
  parameter: V1Parameter;
  schema: yup.BaseSchema;
}): ReactElement {
  const { parameter, schema } = props;
  assertDefined(parameter.Name); // swagger def is too permissive, it must be defined

  const required = !parameter.Optional;

  if (parameter.FromFile) {
    return (
      <FileUploadFormField
        name={`Parameters.${parameter.Name}`}
        label={getFormLabelFromParameter(parameter)}
        helperText={parameter.Help}
        required={required}
      />
    );
  }

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
): yup.BaseSchema {
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
    Parameters: yup.object().shape(parameterSchemas),
  });

  const initialParameterValues = createInitialParameterValues(flavorParameters);

  const { user } = useUserAuth();
  initialParameterValues.name = generateClusterName(user?.Name || '');

  const initialValues: FormikValues = {
    ID: flavorId,
    Description: '',
    Parameters: initialParameterValues,
  };

  const [error, setError] = useState<{
    message?: string;
    response?: {
      data?: {
        error?: string;
      };
    };
  }>();

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
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
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
