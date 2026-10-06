# SPEC-B · 前端（执行者：glm5.3-flash · 分支 feat/frontend）

> 开工前必读：仓库根 `AGENTS.md`、`.agents/skills/dev-conventions.md`（技能索引，**每个任务开做前先读对应技能**）、`.agents/skills/provider-streaming/SKILL.md`（B3 核心）。
> 本 spec 已写到"照做即可"粒度：页面结构、组件清单、store 形态都已钦点，按序执行，不自由发挥目录/命名。

## 边界与纪律

- 只写 `apps/desktop/frontend/src/{app,components,features}/**`
- **禁止 import `bindings/**` 与 `@wailsio/runtime`**（eslint 强制）；一切 Go 能力经 `await desktop()`（`@/shared/desktop/adapter`）
- 类型一律从 `@contracts/index` import，**禁止在前端重复定义 contracts 已有类型**
- 调用 Go 之前先看 mock：`src/shared/desktop/mock.ts` 有全部行为的假实现，B 阶段开发不依赖后端进度
- 每个任务完成跑：`bun run lint && bun run typecheck && bun run test`（frontend 目录内）
- shadcn-vue 组件已全量在 `src/components/ui/`，**只 import 不改写**；设计 token 见 `src/app/style.css`，禁止硬编码颜色

## 通用调用模式（先记住这个）

```ts
import { desktop } from '@/shared/desktop/adapter';

const bridge = await desktop();           // 模块/store 初始化时调用一次并持有
const services = await bridge.listServices();
```

## B1 应用骨架

1. **shell 布局** `src/shell/layout/app-shell.vue`：左侧导航（宽 200px，`bg-sidebar`）+ 右侧 `<RouterView>`。导航项（router-link，激活态 `bg-state-selected`）：翻译 `/`（Languages 图标）、服务 `/services`（Plug）、历史 `/history`（History）、设置 `/settings`（Settings）。图标从 `@lucide/vue` import
2. **路由** 改 `src/app/router/index.ts`：`/` `/services` `/history` `/settings` 四个主页面去掉占位指向真实页面；保留 `/popup` `/input`（B5 用）
3. **入口装配** `src/app/App.vue`：`/popup` `/input` 路由直接渲染（无 shell）；其余包在 `app-shell` 内。实现方式：`app-shell.vue` 里 `<RouterView>`，两个独立路由组件（已存在）不需要 shell —— 用路由嵌套（父路由 component=app-shell，children 四页）
4. **Pinia stores**（全部 setup store 风格，见 pinia-store 技能）：
   - `src/features/services/store.ts`：`services: Ref<ServiceSetting[]>`、`load()`、`toggle(id)`、`save(s)`、`remove(id)`
   - `src/features/translate/store.ts`：见 B3
   - `src/features/history/store.ts`：`items/total/query`、`load()`、`toggleFavorite(item)`、`remove(id)`、`clearAll()`
   - `src/features/settings/store.ts`：`settings: Ref<AppSettings | null>`、`load()`、`save()`

验收：四个页面可导航、亮暗主题跟随系统（`html.dark` class 由 `@vueuse/core` 的 `usePreferredDark` 初始化——settings.theme==='system' 时）。

## B2 服务配置页（`src/features/services/pages/index.vue`）

布局：页头（标题"翻译服务" + `Button`「添加服务」）+ 服务卡片列表（每项 `Card`：名称、类型 `Badge`、`Switch` 启用、编辑/删除 `DropdownMenu`）。

- **编辑弹窗** `components/service-dialog.vue`（`Dialog` + `Form`/`Input`/`Label`/`Select`）：字段按类型显隐——公共：name、type（Select：llm/deepl/google/microsoft）；仅 llm 显示：baseUrl、apiKey（`type=password`）、model、authScheme（Select：bearer/x-api-key）。校验用 `zod`（name 必填；llm 时 baseUrl 必填）
- **连通性测试**：弹窗内 `Button`「测试」→ `bridge.testService(service)` → loading（`Spinner`）→ 成功显示样例译文与延迟，失败显示 error（`Alert` destructive）
- 新增：id 空串传给 store.save；排序 order 取当前最大 +1

验收：mock 模式下增删改查 + 测试全流程可走；vitest 覆盖 store 的 save/toggle/remove（mock bridge 驱动）。

## B3 翻译页（`src/features/translate/pages/index.vue`，核心页面）

布局：上方输入区 + 下方结果区。

- **输入区**：`Textarea`（3 行，自动聚焦）+ 语言选择（源 `Select` 默认 auto + 目标 `Select` 默认 zh-Hans，选项来自 `LANGUAGES` 常量，标签用 `languageLabel(code)`）+ 「翻译」`Button`（Enter 触发）+「清空」
- **结果区（多开对照）**：启用服务数 1 → 单卡片；≥2 → `Tabs`（每个服务一个 tab，标签=服务名）；每个 tab 内：状态区（pending=`Spinner`、streaming=文本渐进 + 末尾闪烁光标动画 `animate-pulse`、retrying=`Badge`「重试中 第N次」、error 按 errorCode 分档文案 `Alert`）+ 译文区（`whitespace-pre-wrap`，结果可滚动）+ 操作条（复制=`Button` variant=ghost 调 `navigator.clipboard.writeText` + 重译 + 收藏）
- **translate store**（严格按 provider-streaming 技能）：
  ```ts
  // state: requestId、Map<serviceId, {status, text, error}>、input、sourceLang、targetLang
  // actions: translate()（清空旧状态 → bridge.translate → 订阅 onTranslateEvent 聚合）
  //          cancel()（bridge.cancelTranslate）、retry(serviceId)
  // finish 事件以 e.text 覆盖累积文本；cancelled 不算失败
  ```
- 翻译完成（首个服务 finish 后）调 `bridge.saveHistory` 存历史（serviceName 冗余存）

验收：mock 模式流式渲染流畅可见（30ms/片渐进）；vitest 覆盖 store 的事件聚合（用 mock bridge 驱动，断言 (requestId,serviceId) 聚合与 finish 覆盖行为）。

## B4 历史页（`src/features/history/pages/index.vue`）

- 顶部过滤：`Input`（关键词，防抖 300ms `@vueuse/core` 的 `refDebounced`）+ 服务 `Select` + 收藏 `Checkbox`/`Switch` + 「清空历史」`Button` variant=destructive（`AlertDialog` 二次确认）
- 列表 `Table`：时间（相对时间，手写 60 行内的 formatRelative）、原文（截断 60 字）、译文（截断）、服务名 `Badge`、收藏 `Button`（Star/StarOff 图标切换）、操作（查看详情 `Dialog` 全文 + 重新翻译=跳翻译页带入文本、删除）
- 分页：`limit=50`，滚动到底加载下一页（offset 递增）

验收：mock 驱动增查删收全流程；vitest 覆盖 store 过滤参数拼装。

## B5 浮窗与输入窗

- **popup**（`src/app/pages/popup-page.vue` 重写）：紧凑结果面板——外层 `rounded-xl border bg-card shadow-lg p-4`（透明窗口内的卡片本体）；顶部：服务切换（`Segmented` 或小 `Tabs`）+ 操作（复制/钉住 `Pin`/关闭 `X`——钉住与关闭仅 UI 态，窗口控制后续主线接 wails API）；中部译文（复用 B3 的状态渲染逻辑，**抽成 `components/stream-result.vue` 共享**）
- **input**（`src/app/pages/input-page.vue` 重写）：顶部输入框 `Input`（autofocus，Enter 翻译）+ 语言对（紧凑 Select）+ 下方复用 `stream-result.vue`
- 两个窗口独立初始化自己的 store 实例（同 SPA 不同窗口，Pinia 各自独立，无需共享）

验收：浏览器直开 `http://localhost:9245/#/popup` 可见完整 UI 并 mock 翻译。

## B6 设置页（`src/features/settings/pages/index.vue`）

- 通用：默认源/目标语言 `Select`（写回 settings）
- 快捷键：展示 `HotkeyConfig` 两项（`Kbd` 组件渲染键位）；「录制」按钮捕获 keydown 显示（`event.altKey` 等 → 'Alt+D' 格式），保存写回 settings
- 外观：主题 `Select`（light/dark/system）→ 切 `html` 的 `.dark` class
- 关于：版本（`__APP_VERSION__` 若无则硬编码 0.1.0）、`bridge.licenseInfo()` 显示 plan

验收：设置读写 mock 全流程；主题切换即时生效。

## 任务内顺序与全局验收

```
B1 → B2 → B3 → B4/B5/B6（后三个任意顺序）
```

全局验收（B 阶段收尾）：

```bash
cd apps/desktop/frontend
bun run lint && bun run typecheck && bun run test   # 三绿
bun run dev   # 浏览器直开，四个页面 + #/popup #/input 全 UI 流程可走（mock）
```

完成后通知主线合入 `feat/frontend`，真实链路联调（I1）由主线统筹。
