// eslint-disable-next-line @typescript-eslint/no-var-requires
const { createProxyMiddleware } = require('http-proxy-middleware');

const defaultOptions = {
  target: process.env.DEV_INFRA_API_ENDPOINT || 'https://localhost:8443',
  changeOrigin: true,
  secure: false,
};

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
module.exports = function main(app) {
  app.use('/v1', createProxyMiddleware(defaultOptions));
};
