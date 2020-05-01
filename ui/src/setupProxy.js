/* eslint @typescript-eslint/no-var-requires: 0, @typescript-eslint/explicit-function-return-type: 0 */

const { createProxyMiddleware } = require('http-proxy-middleware');

const defaultOptions = {
  target: process.env.INFRA_API_ENDPOINT || 'https://dev.infra.stackrox.com',
  changeOrigin: true,
  secure: false,
};

module.exports = function main(app) {
  app.use('/v1', createProxyMiddleware(defaultOptions));
};
