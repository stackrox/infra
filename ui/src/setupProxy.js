/* eslint @typescript-eslint/no-var-requires: 0, @typescript-eslint/explicit-function-return-type: 0 */

const { createProxyMiddleware } = require('http-proxy-middleware');

const defaultOptions = {
  target: process.env.DEV_INFRA_API_ENDPOINT || 'https://localhost:8443',
  changeOrigin: true,
  secure: false,
};

module.exports = function main(app) {
  app.use('/v1', createProxyMiddleware(defaultOptions));
};
