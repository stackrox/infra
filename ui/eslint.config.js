import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import prettier from "eslint-plugin-prettier";
import tsParser from "@typescript-eslint/parser";
import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

export default [{
    ignores: [
        "node_modules",
        "build",
        "coverage",
        "src/generated",
        "tailwind-plugins",
        "!**/.eslintrc.js",
        "!**/.prettierrc.js",
        "!**/.postcssrc.js",
    ],
}, ...fixupConfigRules(compat.extends(
    "react-app",
    "plugin:@typescript-eslint/recommended",
    "plugin:@typescript-eslint/recommended-requiring-type-checking",
    "plugin:eslint-comments/recommended",
    "airbnb-typescript",
    "prettier",
)), {
    plugins: {
        "@typescript-eslint": fixupPluginRules(typescriptEslint),
        prettier,
    },

    languageOptions: {
        parser: tsParser,
        ecmaVersion: 5,
        sourceType: "script",

        parserOptions: {
            project: "./tsconfig.eslint.json",
            tsconfigRootDir: "/Users/house/dev/stack/infra/ui",
        },
    },

    rules: {
        "prettier/prettier": "error",
        "react/prop-types": "off",
        "react/require-default-props": "off",

        "@typescript-eslint/no-unused-vars": ["error", {
            argsIgnorePattern: "^_",
        }],

        "@typescript-eslint/no-unsafe-assignment": "off",
        "@typescript-eslint/no-unsafe-call": "off",

        "import/no-extraneous-dependencies": ["error", {
            devDependencies: [
                "**/.eslintrc.js",
                "tailwind.config.js",
                ".prettierrc.js",
                ".postcssrc.js",
                "src/setupProxy.js",
                "src/setupTests.ts",
                "**/*.test.tsx",
            ],

            optionalDependencies: false,
        }],
    },
}, {
    files: ["src/**/*"],

    languageOptions: {
        globals: {
            ...globals.browser,
        },
    },
}, {
    files: ["**/*.test.ts"],

    languageOptions: {
        globals: {
            ...globals.browser,
            ...globals.jest,
        },
    },
}];