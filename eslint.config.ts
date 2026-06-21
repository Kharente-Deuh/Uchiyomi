// @ts-check
import type { OptionsConfig, TypedFlatConfigItem } from '@antfu/eslint-config'
import antfu from '@antfu/eslint-config'
import nuxt from './.nuxt/eslint.config.mjs'

const customCfg: TypedFlatConfigItem = {
  rules: {
    'pnpm/yaml-enforce-settings': 'off',
    'unicorn/no-null': 'off',
    'ts/explicit-function-return-type': ['error', { allowExpressions: true }],
    'no-unused-private-class-members': 'error',
    'no-unused-expressions': 'error',
    'no-implicit-globals': 'error',
    'unicorn/no-unused-properties': 'error',
    'padding-line-between-statements': [
      'error',
      { blankLine: 'always', prev: 'block-like', next: '*' },
      { blankLine: 'always', prev: '*', next: 'return' },
      { blankLine: 'always', prev: '*', next: 'throw' },
    ],
    'unicorn/prefer-ternary': 'off',
    'no-useless-call': 'error',
    'unused-imports/no-unused-imports': 'error',
    'style/brace-style': ['error', '1tbs'],
    'style/dot-location': ['error', 'property'],
    'style/implicit-arrow-linebreak': ['error', 'beside'],
    'style/no-extra-parens': 'error',
    'style/quote-props': ['error', 'consistent-as-needed', { unnecessary: true }],
    'unicorn/prevent-abbreviations': 'off',
    'unicorn/switch-case-braces': 'off',
    'unicorn/catch-error-name': 'off',
    'unicorn/no-static-only-class': 'off',
    'unicorn/filename-case': 'off',
    'ts/explicit-module-boundary-types': 'off',
    'unicorn/expiring-todo-comments': 'off',
    'style/quotes': ['error', 'single', { allowTemplateLiterals: 'always', avoidEscape: true }],
    'curly': ['error', 'all'],
    'node/prefer-global/process': 'off',
    'vue/custom-event-name-casing': 'off',
    'vue/component-options-name-casing': ['error', 'PascalCase'],
    'ts/consistent-type-definitions': 'off',
    'vue/first-attribute-linebreak': ['error', {
      singleline: 'beside',
      multiline: 'below',
    }],
    'no-warning-comments': ['warn', { terms: ['todo', 'fixme'], location: 'anywhere' }],
    'vue/max-attributes-per-line': ['error', {
      singleline: {
        max: 2,
      },
      multiline: {
        max: 1,
      },
    }],
  },
}

const opts: OptionsConfig & Omit<TypedFlatConfigItem, 'files' | 'ignores'> = {
  ignores: ['prisma/generated', 'docs', '.nuxt', '.output', 'dist', '**/*.md'],
  vue: true,
  typescript: true,
  formatters: true,
  pnpm: true,
  unicorn: { allRecommended: true },
  lessOpinionated: false,
  gitignore: true,
  stylistic: {
    indent: 2,
    quotes: 'single',
    semi: false,
    jsx: false,
  },
  regexp: true,
  autoRenamePlugins: true,
}

export default antfu(opts, customCfg).append(nuxt())
