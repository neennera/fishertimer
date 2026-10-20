import babelParser from "@babel/eslint-parser";
import js from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import turboPlugin from "eslint-plugin-turbo";
import onlyWarn from "eslint-plugin-only-warn";

/**
 * A shared ESLint configuration for the repository.
 *
 * @type {import("eslint").Linter.Config[]}
 * */
export const config = [
  // ESLint only lints .js/.mjs/.cjs unless a config names other files, so
  // without this every .ts/.tsx file was skipped as "no matching configuration".
  { files: ["**/*.{js,mjs,cjs,jsx,ts,tsx,mts,cts}"] },
  js.configs.recommended,
  eslintConfigPrettier,
  {
    languageOptions: {
      parser: babelParser,
      parserOptions: {
        requireConfigFile: false,
        babelOptions: {
          presets: ["@babel/preset-typescript"],
          // Babel 8 no longer parses JSX in .tsx on its own. Only .jsx/.tsx get
          // it, so `<T>value` casts in .ts files still parse.
          overrides: [{ test: /\.[jt]sx$/, parserOpts: { plugins: ["jsx"] } }],
        },
      },
    },
    plugins: {
      turbo: turboPlugin,
    },
    rules: {
      "turbo/no-undeclared-env-vars": "warn",
    },
  },
  {
    // The babel parser hands ESLint TypeScript without type information, so
    // these core rules misfire on types (every type name is "undefined",
    // `{ children }: { children: ReactNode }` is a "duplicate param"). tsc
    // checks all of them already. Same list as typescript-eslint's
    // eslint-recommended config, plus no-unused-vars; typescript-eslint
    // itself does not support TypeScript 7 yet.
    files: ["**/*.{ts,tsx,mts,cts}"],
    rules: {
      "constructor-super": "off",
      "getter-return": "off",
      "no-class-assign": "off",
      "no-const-assign": "off",
      "no-dupe-args": "off",
      "no-dupe-class-members": "off",
      "no-dupe-keys": "off",
      "no-func-assign": "off",
      "no-import-assign": "off",
      "no-new-native-nonconstructor": "off",
      "no-obj-calls": "off",
      "no-redeclare": "off",
      "no-setter-return": "off",
      "no-this-before-super": "off",
      "no-undef": "off",
      "no-unreachable": "off",
      "no-unsafe-negation": "off",
      "no-unused-vars": "off",
    },
  },
  {
    plugins: {
      onlyWarn,
    },
  },
  {
    ignores: ["dist/**"],
  },
];
