module.exports = {
  extends: ['prettier/react'],
  env: {
    browser: true,
  },
  rules: {
    // will rely on TypeScript compile time checks instead
    'react/prop-types': 'off',

    'import/no-extraneous-dependencies': [
      'error',
      {
        devDependencies: ['**/*.test.tsx', 'src/setupTests.ts', 'src/setupProxy.js'],
        optionalDependencies: false,
      },
    ],
  },
  overrides: [
    {
      files: ['*.test.ts'],
      env: {
        browser: true,
        jest: true,
      },
    },
  ],
};
