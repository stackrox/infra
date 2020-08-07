/* eslint @typescript-eslint/no-var-requires: 0 */

const baseConfig = require('@stackrox/tailwind-config');

module.exports = {
  ...baseConfig,
  purge: ['./public/index.html', './src/**/*.tsx', './src/**/*.tw.css'],
};
