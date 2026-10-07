import globals from 'globals';
import pluginJs from '@eslint/js';
import tseslint from 'typescript-eslint';
import pluginVue from 'eslint-plugin-vue';
import pluginPrettierRecommendedConfigs from 'eslint-plugin-prettier/recommended';
import pluginBoundaries from 'eslint-plugin-boundaries';
import parserVue from 'vue-eslint-parser';

// ============================================================================
// Tako 前端 lint 基线（改写自上游前端 eslint 基线）
// 裁剪：design-token 自定义插件、vuejs-accessibility、存量棘轮（新仓库零存量）。
// 保留：boundaries 分层、token 色彩护栏、max-lines 体量门槛、wails 触点红线。
// ============================================================================

export default [
  // 全局忽略: 构建产物与 wails 生成物
  {
    ignores: ['dist/**', 'node_modules/**', 'bindings/**', 'public/**'],
  },
  // eslint 默认推荐规则
  pluginJs.configs.recommended,
  // ts 默认推荐规则
  ...tseslint.configs.recommended,
  // `_` 前缀是有意忽略的参数/变量约定
  {
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_', caughtErrorsIgnorePattern: '^_' },
      ],
    },
  },
  // vue3 基础推荐规则
  ...pluginVue.configs['flat/recommended'],
  // 未注册组件护栏: 模板使用但未 import 的组件静默降级为原生标签
  {
    files: ['src/**/*.vue'],
    rules: {
      'vue/no-undef-components': [
        'error',
        {
          ignorePatterns: ['RouterView', 'RouterLink', 'router-view', 'router-link'],
        },
      ],
    },
  },
  // prettier 默认推荐规则
  pluginPrettierRecommendedConfigs,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2020,
      },
      ecmaVersion: 2020,
      parser: parserVue,
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  // 页面组件为 kebab-case 单词文件名，关闭多词强制
  {
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
  // 测试文件允许内联多组件挂载
  {
    files: ['**/*.spec.ts', '**/*.spec.tsx'],
    rules: {
      'vue/one-component-per-file': 'off',
    },
  },
  // registry 分发组件库 (shadcn-vue 复制式): 不手工维护代码质量, 存量告警降级
  {
    files: ['src/components/ui/**'],
    rules: {
      '@typescript-eslint/no-empty-object-type': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      'no-undef': 'off',
      'vue/require-default-prop': 'off',
    },
  },
  // ============================================================================
  // 架构边界护栏 (feature-first, 见 docs/agents)
  // ============================================================================
  {
    files: ['src/**/*'],
    plugins: {
      boundaries: pluginBoundaries,
    },
    settings: {
      'boundaries/elements': [
        { type: 'app', pattern: 'src/app/**' },
        { type: 'shell', pattern: 'src/shell/**' },
        { type: 'shared', pattern: 'src/shared/**' },
        { type: 'feature', pattern: 'src/features/*/**', capture: ['featureName'] },
        { type: 'component', pattern: 'src/components/**' },
      ],
      'import/resolver': {
        typescript: {
          alwaysTryTypes: true,
          project: ['./tsconfig.json', './tsconfig.app.json'],
        },
      },
    },
    rules: {
      'boundaries/dependencies': [
        'error',
        {
          default: 'allow',
          policies: [
            // features 之间禁止互相导入
            {
              from: { element: { type: 'feature' } },
              disallow: {
                to: {
                  element: {
                    type: 'feature',
                    captured: { featureName: '!{{from.element.captured.featureName}}' },
                  },
                },
              },
              message: 'features 之间禁止互相导入, 跨 feature 复用的代码上移 src/shared/',
            },
            // features 禁止反向依赖应用壳
            {
              from: { element: { type: 'feature' } },
              disallow: { to: { element: { types: { anyOf: ['app', 'shell'] } } } },
              message: 'feature 不允许导入 app/shell (依赖方向: app -> feature)',
            },
            // shared 不感知任何业务 feature
            {
              from: { element: { type: 'shared' } },
              disallow: { to: { element: { type: 'feature' } } },
              message: 'shared 不允许导入 features (依赖方向: feature -> shared)',
            },
          ],
        },
      ],
    },
  },
  // ============================================================================
  // wails 触点红线: bindings / @wailsio/runtime 只允许出现在 adapter
  // ============================================================================
  {
    files: ['src/**/*'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@wailsio/runtime', '@wailsio/runtime/*', '../../bindings/**', '*bindings/tako/**'],
              message: 'wails 能力一律经 @/shared/desktop/adapter (desktop()), 禁止直接 import',
            },
          ],
        },
      ],
    },
  },
  // 例外: adapter 是 wails 的唯一触点
  {
    files: ['src/shared/desktop/**'],
    rules: {
      'no-restricted-imports': 'off',
    },
  },
  // ============================================================================
  // 设计 token 护栏 (无自定义插件版): 业务界面禁止硬编码颜色, 一律走 token.
  // ============================================================================
  {
    files: [
      'src/features/**/*.vue',
      'src/shell/**/*.vue',
      'src/shared/**/*.vue',
      'src/app/**/*.vue',
      'src/components/**/*.vue',
    ],
    ignores: ['src/components/ui/**'],
    rules: {
      'vue/no-restricted-class': [
        'error',
        // Tailwind 原生色板裸用: 中性色走 foreground/muted-foreground, 状态色走 token
        '/^(?:[^:\\s]+:)*(bg|text|border|ring|fill|stroke|from|via|to|divide|outline|decoration|accent|caret|shadow)-(slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\\d+(\\/[\\d.]+|\\/\\[[^\\]]+\\])?$/',
        // 任意 hex: 使用 style.css 中的 CSS 变量/token
        '/\\[#[0-9a-fA-F]{3,8}\\]/',
        // dark 模式手工黑白双写: 用语义 token 自动适配主题
        '/dark:.*(bg|text|border)-(white|black)$/',
      ],
    },
  },
  // ============================================================================
  // 文件体量门槛: 新文件硬上限 500 有效行 (skipComments+skipBlankLines)
  // ============================================================================
  {
    files: ['src/**/*.{ts,vue}'],
    ignores: ['src/components/ui/**', '**/*.spec.ts'],
    rules: {
      'max-lines': ['error', { max: 500, skipBlankLines: true, skipComments: true }],
    },
  },
];
