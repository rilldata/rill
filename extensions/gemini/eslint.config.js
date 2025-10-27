// @ts-check
import tsEslint from 'typescript-eslint';
import rootConfig from '../../eslint.config.js';

export default [
  ...rootConfig,
  {
    ignores: ['dist/**', 'node_modules/**', 'reports/**', '*.config.js'],
  },
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
  {
    files: ['*.js', '*.config.js'],
    ...tsEslint.configs.disableTypeChecked,
  },
];
