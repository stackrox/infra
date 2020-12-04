/* eslint @typescript-eslint/no-var-requires: 0, @typescript-eslint/explicit-function-return-type: 0 */

const { createProxyMiddleware } = require('http-proxy-middleware');

const defaultOptions = {
  target: process.env.INFRA_API_ENDPOINT || 'https://dev.infra.rox.systems',
  changeOrigin: true,
  secure: false,
};

/**
 * @param {Object} app - Express.js application
 */
module.exports = function main(app) {
  /* eslint-disable @typescript-eslint/no-unsafe-call */
  const proxy = createProxyMiddleware(defaultOptions);
  app.use('/v1', proxy);
  app.use('/login', proxy);
  app.use('/callback', proxy);
  app.use('/logout', proxy);
  app.use('/downloads/infractl-*', proxy);
  /* eslint-enable @typescript-eslint/no-unsafe-call */
};
