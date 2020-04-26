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
  ],
  parserOptions: {
    project: './tsconfig.eslint.json',
    tsconfigRootDir: __dirname,
  },
  rules: {
    'prettier/prettier': 'error',

    'import/no-extraneous-dependencies': [
      'error',
      {
        devDependencies: [
          '**/.eslintrc.js',
          'tailwind.config.js',
          '.prettierrc.js',
          '.postcssrc.js',
        ],
        optionalDependencies: false,
      },
    ],
  },
};
