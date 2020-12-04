/* eslint @typescript-eslint/no-var-requires: 0 */

// eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
const baseConfig = require('@stackrox/tailwind-config');

// eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
module.exports = {
  ...baseConfig,
  purge: ['./public/index.html', './src/**/*.tsx', './src/**/*.tw.css'],
};
