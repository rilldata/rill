// //@ts-check

import js from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import globals from "globals";
import tsEslint from "typescript-eslint";
import eslintPluginSvelte from "eslint-plugin-svelte";
import vitest from "eslint-plugin-vitest";
import playwright from "eslint-plugin-playwright";

export default [
  js.configs.recommended,
  ...tsEslint.configs.recommended,
  ...eslintPluginSvelte.configs["flat/recommended"],
  eslintConfigPrettier,
  {
    ...playwright.configs["flat/playwright"],
    files: ["**/tests/**"],
  },
  vitest.configs.recommended,
  ...eslintPluginSvelte.configs["flat/prettier"],
  {
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: { ...globals.node, ...globals.browser },
      parserOptions: {
        project: true,
        tsconfigRootDir: import.meta.dirname,
        parser: tsEslint.parser,
        extraFileExtensions: [".svelte"],
      },
    },
    rules: {
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/ban-ts-comment": "warn",
      "@typescript-eslint/no-unsafe-assignment": "warn",
      "@typescript-eslint/no-unsafe-member-access": "warn",
      "@typescript-eslint/no-unsafe-argument": "warn",
      "@typescript-eslint/no-unsafe-call": "warn",
      "@typescript-eslint/no-unsafe-return": "warn",
      "@typescript-eslint/no-floating-promises": "warn",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",
      "@typescript-eslint/unbound-method": "warn",
      "@typescript-eslint/require-await": "warn",
      "@typescript-eslint/restrict-template-expressions": "warn",
      "@typescript-eslint/no-redundant-type-constituents": "warn",
      "@typescript-eslint/no-unsafe-enum-comparison": "warn",
      "@typescript-eslint/no-misused-promises": "warn",
      "@typescript-eslint/no-duplicate-enum-values": "warn",
      "@typescript-eslint/await-thenable": "warn",
      "@typescript-eslint/no-implied-eval": "warn",
      "@typescript-eslint/no-base-to-string": "warn",
    },
  },
  {
    ignores: [
      "**/postcss.config.cjs",
      "web-local/build/*",
      "web-common/build/*",
      "web-admin/build/*",
      "web-common/src/components/modal/*.js",
      "**/.svelte-kit/",
      "**/node_modules",
      "**/playwright.config.js",
      "**/svelte.config.js",
    ],
  },
  {
    files: ["*.js"],
    ...tsEslint.configs.disableTypeChecked,
  },
  {
    files: ["**/*.svelte"],
    rules: {
      "svelte/no-target-blank": "warn",
      "svelte/no-at-html-tags": "error",
      "svelte/no-at-debug-tags": "error",
      "svelte/require-each-key": "warn",
      "svelte/prefer-destructured-store-props": "warn",
      "svelte/require-optimized-style-attribute": "warn",
      "svelte/prefer-class-directive": "warn",
      "svelte/require-store-reactive-access": "warn",
      "svelte/valid-prop-names-in-kit-pages": "warn",
      //   "svelte/require-event-dispatcher-types": "warn",
      //   "svelte/sort-attributes": "warn",
    },
  },
];
