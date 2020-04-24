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
        devDependencies: ['**/*.test.tsx', 'src/setupTests.ts', 'src/setupProxy.ts'],
        optionalDependencies: false,
      },
    ],
  },
  overrides: [
    {
      files: ['*.test.js'],
      env: {
        browser: true,
        jest: true,
      },
    },
  ],
};
