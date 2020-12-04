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

    'import/no-extraneous-dependencies': [
      'error',
      {
        devDependencies: [
          '**/.eslintrc.js',
          'tailwind.config.js',
          '.prettierrc.js',
          '.postcssrc.js',
          'src/setupProxy.js',
          'src/setupTests.ts',
          '**/*.test.tsx',
        ],
        optionalDependencies: false,
      },
    ],
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
