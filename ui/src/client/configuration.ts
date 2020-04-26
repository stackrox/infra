import { Configuration, ConfigurationParameters } from 'generated/client';

const parameters: ConfigurationParameters = {
  basePath: `${window.location.protocol}//${window.location.host}`,
};

export default new Configuration(parameters);
