module.exports = function override(config) {
  const fallback = config.resolve.fallback || {};
  Object.assign(fallback, {
    crypto: false,
    stream: false,
    buffer: false,
    path: false,
    fs: false,
  });
  config.resolve.fallback = fallback;
  return config;
};
