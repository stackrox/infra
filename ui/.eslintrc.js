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

    // will rely on TypeScript compile time checks instead
    'react/prop-types': 'off',
    // allow ordered declaration of components
    '@typescript-eslint/no-use-before-define': 'off',
    // function component default props are OK
    'react/require-default-props': 'off',
    // proto definitions mean that we have typed undefined attributes that are always defined
    // e.g. cluster.ID
    '@typescript-eslint/restrict-template-expressions': 'off',

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
