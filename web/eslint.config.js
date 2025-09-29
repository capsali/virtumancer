import { defineConfig, globalIgnores } from 'eslint/config'
import globals from 'globals'
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import pluginVitest from '@vitest/eslint-plugin'
import pluginPlaywright from 'eslint-plugin-playwright'
import tsParser from '@typescript-eslint/parser'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

export default defineConfig([
  {
    name: 'app/files-to-lint',
    files: ['**/*.{js,mjs,jsx,ts,tsx,vue}'],
  },

  globalIgnores([
    '**/dist/**', 
    '**/dist-ssr/**', 
    '**/coverage/**',
    '**/.vite/**',
    '**/public/spice/**',
    '**/_archived_src_*/**'
  ]),

  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },

  js.configs.recommended,
  
  // Vue configuration with TypeScript support
  ...pluginVue.configs['flat/essential'],
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    rules: {
      // Disable problematic rules for Vue TypeScript files
      'no-unused-vars': 'off',
      'no-undef': 'off',
    },
  },

  // TypeScript files configuration
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    rules: {
      'no-unused-vars': 'off',
      'no-undef': 'off',
    },
  },

  // Test files configuration
  {
    files: ['**/*.test.{js,mjs,jsx,ts,tsx,vue}'],
    ...pluginVitest.configs.recommended,
  },

  {
    files: ['e2e/**/*.{js,mjs,jsx,ts,tsx,vue}'],
    ...pluginPlaywright.configs['flat/recommended'],
  },

  skipFormatting,
])
