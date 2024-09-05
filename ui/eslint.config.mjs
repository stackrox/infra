// @ts-check

import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import prettier from 'eslint-plugin-prettier';
import jsxa11y from 'eslint-plugin-jsx-a11y';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {  ignores: [
      'node_modules/**',
      'build/**',
      'src/generated/**',
      '!**/.prettierrc.js',
      '!**/.postcssrc.js',
    ],},
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    plugins: {
      prettier,
      'jsx-a11y': jsxa11y,
    },
    rules: {
      'prettier/prettier': 'warn',

      "react/prop-types": "off",
      "react/require-default-props": "off",

      "jsx-a11y/label-has-associated-control": "warn",

      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-require-imports": "off",
      "no-undef": "off",
    },
  },
);

