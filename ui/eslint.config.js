import { fixupConfigRules, fixupPluginRules } from '@eslint/compat';
import typescriptEslint from '@typescript-eslint/eslint-plugin';
import prettier from 'eslint-plugin-prettier';
import tsParser from '@typescript-eslint/parser';
import globals from 'globals';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import js from '@eslint/js';
import { FlatCompat } from '@eslint/eslintrc';

const filename = fileURLToPath(import.meta.url);
const dirname = path.dirname(filename);
const compat = new FlatCompat({
  baseDirectory: dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  {
    files: ['**/*.js', '**/*.jsx', '**/*.ts', '**/*.tsx'],
    ignores: [
      'node_modules',
      'build',
      'coverage',
      'src/generated',
      'tailwind-plugins',
      '!**/.eslintrc.js',
      '!**/.prettierrc.js',
      '!**/.postcssrc.js',
    ],
  },
  ...fixupConfigRules(
    compat.extends(
      'plugin:@typescript-eslint/recommended',
      'plugin:@typescript-eslint/recommended-requiring-type-checking',
      'plugin:eslint-comments/recommended',
      'airbnb-typescript',
      'prettier',
    ),
  ),
  {
    plugins: {
      '@typescript-eslint': fixupPluginRules(typescriptEslint),
      prettier,
    },

    languageOptions: {
      parser: tsParser,
      ecmaVersion: 5,
      sourceType: 'script',

      parserOptions: {
        project: './tsconfig.eslint.json',
      },
    },

    rules: {
      'prettier/prettier': 'error',
      'react/prop-types': 'off',
      'react/require-default-props': 'off',

      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',
      '@typescript-eslint/no-floating-promises': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/lines-between-class-members': 'off',
      '@typescript-eslint/no-throw-literal': 'off',
      '@typescript-eslint/no-require-imports': 'off',
      'eslint-comments/disable-enable-pair': 'off',
      'eslint-comments/no-unlimited-disable': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/naming-convention': 'off',

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
  },
  {
    files: ['src/**/*'],

    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },
  {
    files: ['**/*.test.ts'],

    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.jest,
      },
    },
  },
];
