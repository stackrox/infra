module.exports = {
  plugins: ['@typescript-eslint', 'prettier'],
  extends: [
    'react-app',
    'airbnb-typescript',
    'plugin:@typescript-eslint/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking',
    'plugin:eslint-comments/recommended',
    'prettier',
    'prettier/@typescript-eslint',
  ],
  parserOptions: {
    project: './tsconfig.eslint.json',
    tsconfigRootDir: __dirname,
  },
  rules: {
    'prettier/prettier': [
      'error',
      {
        printWidth: 100,
        singleQuote: true,
      },
    ],
  },
};
