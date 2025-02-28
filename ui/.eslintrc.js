module.exports = {
  plugins: ['@typescript-eslint', 'prettier'],
  parser: '@typescript-eslint/parser',
  extends: [
    'react-app',
    'plugin:@typescript-eslint/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking',
    'plugin:eslint-comments/recommended',
    'airbnb-typescript',
    'prettier',
    'prettier/@typescript-eslint',
    'prettier/react',
  ],
  parserOptions: {
    project: './tsconfig.eslint.json',
    tsconfigRootDir: __dirname,
  },
  rules: {
    'prettier/prettier': 'error',

    // rely on TypeScript compile time definitions & checks instead of propTypes and defaultProps
    'react/prop-types': 'off',
    'react/require-default-props': 'off',

    'no-nested-ternary': 'off',

    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],

    'import/no-extraneous-dependencies': [
      'error',
      {
        devDependencies: [
          '**/.eslintrc.js',
          '.prettierrc.js',
          'src/setupProxy.js',
          'src/setupTests.ts',
          '**/*.test.tsx',
        ],
        optionalDependencies: false,
      },
    ],

    // The following rules began throwing errors after tooling updates. These should probably
    // be re-enabled and fixed at some point
    'react/jsx-no-bind': 'off',
    '@typescript-eslint/no-floating-promises': 'off',
    '@typescript-eslint/no-misused-promises': 'off',
    '@typescript-eslint/no-unsafe-argument': 'off',
    // End disabled tooling rules
  },

  overrides: [
    {
      files: ['src/**/*'],
      env: {
        browser: true,
      },
    },
    {
      files: ['*.test.ts'],
      env: {
        browser: true,
        jest: true,
      },
    },
  ],
};
