// @ts-check

import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import prettier from 'eslint-plugin-prettier';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{js,ts,tsx}'],
    ignores: [
      'node_modules',
      'src/generated',
      '!**/.prettierrc.js',
      '!**/.postcssrc.js',
    ],
    plugins: {
      prettier,
    },
    rules: {
      "react/prop-types": "off",
      "react/require-default-props": "off",
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-require-imports": "off",
      "@typescript-eslint/no-unused-vars": "off",
      "import/prefer-default-export": "off",
      "no-undef": "off",
    },
  },
);

