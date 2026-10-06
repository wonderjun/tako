---
name: vue-component
description: 在 tako 前端 编写或修改业务 .vue 组件（SFC 结构、props/emits、registry 组件消费、交互反馈细则）时使用。
---

# Vue Component Conventions

全部组件用 `<script setup lang="ts">`，样式走 Tailwind 语义 token，**多数组件无 `<style>` 块**。本文骨架提炼自 `src/features/settings/mcp/pages/components/mcp-server-item.vue`（props/emits 与结构正典）。

## 适用场景

- 新建或修改 `features/`、`shell/` 下的业务 `.vue` 组件。
- 决定 props/emits 怎么声明、事件与 toast 怎么反馈。
- 消费 `components/ui/`（shadcn-vue）与 `ai-elements/`（ai-elements-vue）registry 组件。

## SFC 正典骨架

业务组件 template 在前、script 在后（registry 库组件相反，script 在前——不要改它们）：

```vue
<template>
  <div class="flex items-center gap-4 rounded-lg border border-hairline bg-card px-5 py-4">
    <!-- 中文注释标注模板分区 -->
    <span class="truncate text-sm font-medium text-foreground">{{ item.name }}</span>
    <Switch :model-value="item.enabled" @update:model-value="(v: boolean) => emit('toggle', v)" />
    <Button type="button" variant="ghost" size="icon" :disabled="testing" @click="emit('test')">
      <Loader2 v-if="testing" :size="16" class="animate-spin" />
    </Button>
  </div>
</template>

<script setup lang="ts">
// ① vue / 三方（icons 从 lucide-vue-next 单独引）
import { computed } from 'vue';
import { Loader2 } from 'lucide-vue-next';
// ② registry 组件（从 barrel 引，不深引 .vue 文件）
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
// ③ feature 内类型
import type { McpServerListItem } from '@/features/settings/mcp/ai-mcp.type';

// ④ props：类型化泛型，逐 prop 中文 JSDoc
const props = defineProps<{
  item: McpServerListItem;
  /** 是否正在测试该服务器连接. */
  testing?: boolean;
}>();

// ⑤ emits：类型化元组签名，按模板出现顺序排列
const emit = defineEmits<{
  toggle: [enabled: boolean];
  test: [];
}>();

// ⑥ 常量 → computed → watch → 事件处理函数（按模板出现顺序）
/** 工具摘要上限. */
const TOOL_LIMIT = 5;
const toolsSummary = computed(() => props.item.tools?.slice(0, TOOL_LIMIT).map((t) => t.name).join(', ') ?? '');
</script>
```

## 硬性规则

1. **props/emits 一律类型化泛型声明**，禁止 runtime 声明（`defineProps({ foo: String })`）；带默认值用 `withDefaults(defineProps<Props>(), {...})`。
2. 每个非常显自明的 prop 一条中文 JSDoc；`props` 变量在模板用不到时也要保留（script 内经 `props.x` 访问，不解构丢响应性；确需解构用 `toRef`/响应式语法）。
3. emit 事件名用动词（`toggle`、`test`、`delete`）；`update:modelValue` 走 reka 约定。
4. 声明顺序固定：imports → props → emits → 常量 → store/composable 实例化 → ref → computed → watch → function。禁止散乱穿插。
5. 中文 UI 文案内联在模板；映射表用 `Record<K, string>` 常量放 script 顶部。
6. 图标统一 `lucide-vue-next`，尺寸 `:size="16"`、线性 `:stroke-width="1.5"` 与存量一致。

## registry 组件消费

- 只从 barrel import：`import { Button } from '@/components/ui/button'`、ai-elements 同理（`@/components/ai-elements/<域>`）；业务组件**不写 cva、不暴露 variants**、不用 `cn()` 拼变体——需要新变体时改 registry 组件本身。
- **禁止对已存在 registry 组件执行 CLI 覆盖**（`add --overwrite`）；确需升级按 `src/components/CUSTOMIZATIONS.md` 流程：干净工作区 → CLI 覆盖 → `git diff` → 重放 `CUSTOMIZED:` 标记定制 → 更新清单。
- 修改 `components/ui/`、`ai-elements/` 下文件必须在改动处加 `CUSTOMIZED:` 注释并更新 `CUSTOMIZATIONS.md`。
- 新增不存在的组件正常 `bunx shadcn-vue@latest add <name>`。

## 交互反馈细则

- **错误/成功反馈统一 vue-sonner**：`toast.error(msg)` / `toast.success('提示词已复制')`；请求层错误已被拦截器统一 toast，见 `api-query`——组件内不要对同一错误重复 toast。
- **异步删除等破坏性操作禁止用 `AlertDialogAction` 直接触发 mutation**（它会自动关弹窗）：用普通 `Button` 触发 mutation，请求期间 `:disabled`，仅在 `onSuccess` 显式关闭弹窗，失败保持弹窗打开以便重试。
- 可点元素补焦点环：`focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50`。
- 装饰性交互区域写 `aria-label` + `title`（见 mcp-server-item 的图标按钮）。

## a11y 规则

- icon-only 按钮必须提供有意义的 `aria-label`；可点元素同时保留可见焦点环，避免仅依赖图标或颜色传达含义。
- 带 `click` 的非原生交互元素必须提供键盘伴随事件并支持聚焦；优先使用原生 `button`、`a` 或 registry 组件。
- 表单标签应通过 `for`/`id` 关联控件，禁止 `accesskey`、`autofocus` 和无效或冗余 ARIA 属性/角色。
- ESLint 对应规则：`vuejs-accessibility/aria-props`、`vuejs-accessibility/aria-unsupported-elements`、`vuejs-accessibility/no-access-key`、`vuejs-accessibility/no-autofocus`、`vuejs-accessibility/label-has-for`、`vuejs-accessibility/no-redundant-roles`；`vuejs-accessibility/click-events-have-key-events` 与 `vuejs-accessibility/interactive-supports-focus` 用于交互键盘可达性告警。

## 样式规则（与 DESIGN.md 联动）

- 只用语义 token：`bg-primary`、`text-muted-foreground`、`border-hairline`、`text-body`、`shadow-elevation-1`；全表见 `tako 前端/DESIGN.md`。
- eslint `vue/no-restricted-class` 拦截裸色板、hex、`dark:` 双写、任意 `text-[Npx]`；暗色模式由 `.dark` 变量块自动适配，**组件内禁止手写 dark: 变体**。
- 需要新 token 时先改 `src/app/style.css` 的 `@theme inline` 映射，再用类名。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| 组件文件 PascalCase | ~~todo/flash-note/schedule 26 个~~ **已全部整改（2026-08-30，见 feature-organization）** | 历史反例存档 |
| runtime props 声明 / props 无 JSDoc | — | 类型化泛型 + 逐 prop 注释是唯一正典 |
| `AlertDialogAction` 内联异步 mutation | ~~flash-note 标签删除、members API Key 删除、members 退出登录 3 处~~ **已全部整改（2026-08-30）** | 自动关弹窗吞掉失败态，破坏重试回路 |

## 自查清单

- [ ] template 在前 script 在后、声明顺序六段固定？
- [ ] props/emits 类型化泛型 + 中文 JSDoc？
- [ ] registry 组件从 barrel 引、没写 cva/cn？
- [ ] 破坏性异步操作走 Button + onSuccess 关弹窗？
- [ ] 样式全语义 token、可点元素有焦点环与 aria-label？

## 范本文件

- 结构 + props/emits 正典：`src/features/settings/mcp/pages/components/mcp-server-item.vue`
- withDefaults 正典：`src/features/settings/components/settings-section.vue`、`src/features/settings/mcp/pages/components/mcp-modal.vue`
- 常量映射表：`src/shell/layout/components/update-button.vue`（`NOTE_KIND_BADGE_CLASS`）

相关技能：`api-query`、`pinia-store`、`code-style-guide`、`feature-organization`。
