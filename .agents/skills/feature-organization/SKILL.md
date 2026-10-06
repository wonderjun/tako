---
name: feature-organization
description: 在 tako 前端 新建 feature 或子域、决定文件归属位置、整理目录结构或为新页面落位时使用。
---

# Feature Organization

`src/` 按 feature-first 组织，目录即边界（eslint-plugin-boundaries `error` 机器强制）。本文骨架提炼自 `src/features/image-gen/`（小而全正典）与 `src/features/chat/`（全形态正典）。

## 适用场景

- 新建一个 feature（对应一个路由域）或在大 feature 内开子域。
- 拿不准一个新文件该放 feature 根、pages/ 子目录还是 shared/。
- 整理或评审 feature 内部结构。

## 目录正典

形态 A：小而全（image-gen，新 feature 默认照此）：

```
src/features/image-gen/
  image-gen.api.ts        # 该域全部后端接口（函数式，见 api-query）
  image-gen.type.ts       # 该域全部 DTO 类型（见 type-dto）
  stores/image-gen.ts     # 客户端状态（见 pinia-store；钦点目录名 store/，存量 stores/ 待改）
  pages/
    index.vue             # 页面唯一入口，路由只指这里
    composables/          # 页面级 Vue Query hooks 与副作用 hooks
    components/           # 页面私有组件（kebab-case.vue）
```

形态 B：全形态（chat，含流式/常量/工具函数时逐层补齐）：

```
src/features/chat/
  ai-chat.api.ts / ai-chat.type.ts    # feature 根 API 与类型
  mcp-auth.api.ts                     # 按后端契约拆出的独立薄 API
  constants/sse-events.ts             # 前后端契约常量表
  stream/codec.ts                     # 流式协议层（纯函数，见 sse-streaming）
  store/                              # 客户端状态（拆多文件，一文件一职责）
  pages/
    index.vue
    composables/  components/  utils/  constants/
```

规则：

1. **页面入口唯一**：路由（`src/app/router/index.ts`）只指向 `pages/index.vue`；页面级布局文件放 `pages/` 下（如 `pages/space.vue`、`pages/layout.vue`）。既有例外形态（存量，不模仿、新 feature 禁止跟随）：settings 域无 pages 层、路由直挂 `features/settings/index.vue`；flash-note 走 `pages/layout.vue` + `pages/views/*` 子路由；全局错误页在 `app/pages/error/`。
2. **feature 内禁止 index.ts barrel**，全部深路径 import；只有 registry 库（`components/ui/`、`ai-elements/`）有 barrel。
3. **API 与类型在 feature 根**，命名 `<域>.api.ts` / `<域>.type.ts`；次要契约（独立后端 controller）单独一对薄文件（如 `mcp-auth.api.ts`）。
4. **客户端状态目录钦点 `store/`**（单数），文件名 `<职责>.ts`；存量 `stores/` 目录登记反例待改，新代码禁止新建。
5. 所有非组件 `.ts` 文件 kebab-case（`use-task-polling.ts`、`chat-message-normalize.ts`）；`.vue` 组件文件钦点 kebab-case（`result-gallery.vue`）。
6. spec 与被测文件同目录同名：`codec.ts` → `codec.spec.ts`。

## 子域拆分判断

一个大 feature 内出现**独立后端契约 + 独立页面集 + 独立状态**三者齐备的业务块时，开子域目录，子域内重复形态 A：

```
features/settings/mcp/{ai-mcp.api.ts, ai-mcp.type.ts, pages/, stores/}
features/todo/schedule/{schedule.api.ts, pages/, components/, composables/, constants.ts, utils.ts}
```

只有其一（仅复用组件或仅共用类型）不开子域，放 feature 根共享目录（`features/settings/components/`、`navigation.ts`）。

shell 域镜像同一结构（不要在 `components/` 里平铺堆 .vue + hooks 混放）：`src/shell/<域>/` 一级为 `index.vue` + `composables/`（use-*.ts 与 spec）+ `components/`（.vue 与组件配套常量表）；组件域再大时在 `components/` 下开子目录（如 `components/sidebar/` 放 sidebar 系组件，数据 hooks 一律上提 `composables/`，见 `src/shell/layout/`）。

## 文件体量

机器门槛走 eslint `max-lines`（error 级，口径 `skipComments + skipBlankLines`，即去注释去空行后的有效行），本节是人肉判断的补充细则：

- **新文件硬上限 500 有效行，软上限 400 行**——接近软上限就考虑下一档拆分，不要顶格写。
- 拆分触发器：单文件有效行 >500 必拆（lint 直接报错）；组件 template >300 行考虑拆子组件；单个函数 >100 行考虑外移为纯函数/composable。
- 存量超限文件在 `eslint.config.js` 的 `MAX_LINES_RATCHET` 棘轮块逐文件锁死当前有效行数（只许降不许升），拆分达标后删除条目；勿调大数值、勿给新文件加棘轮条目（直接写小）。
- 豁免：registry 复制式组件库（`components/ui/`、`ai-elements/`，不手工维护）、`*.spec.ts`（测试体量是被测对象体量的镜像，拆被测对象时自然分片）。

拆分手法速判（按超限成分对号入座）：

| 超限成分 | 手法 | 去处 |
|---|---|---|
| 可测纯逻辑（转换/解析/派生） | 抽纯函数 + 同名 spec | `pages/utils/` 或子域 `utils*.ts` |
| 响应式状态编排（ref 群 + watch 链） | 抽 composable，去掉页面镜像状态 | `pages/composables/`（三级位置见 composable skill） |
| UI 区块（模板大段/对话框群） | 拆子组件，弹窗 `v-if` 挂载 | `pages/components/` 按业务域子目录 |
| 客户端状态堆积 | store 目录一文件一职责，子系统外移（stream-handlers 先例） | `store/`（见 pinia-store） |
| Wails webview 主进程堆积 | IPC 按域 register 函数 + 窗口壳各自成模块 | `electron/<域>/`（见 `electron/browser/` 外拆先例） |

拆分时保持行为零变化的纪律：先补/确认 spec 锁语义再搬代码；搬移优先整块剪切（不重写）；拆完该文件的棘轮条目随之删除。

## 文件归属速判

| 文件性质 | 位置 |
|---|---|
| 只服务一个页面组件 | `pages/components/` 同级就近 |
| 多页面共用但不出 feature | feature 根 `components/` 或子域 `components/` |
| 纯函数可测逻辑 | feature 内 `utils/`（必须配 spec，见 testing） |
| 被两个以上 feature 引用 | 上移 `src/shared/<域>/`（先问：值不值得抽？能保持无业务语义才上移） |
| 窗口壳/非业务（layout、title-bar、boot-screen） | `src/shell/`，禁止反向被 features 引用 |

## 硬性规则

- `features/A` 禁止 import `features/B`；跨 feature 逻辑上移 `shared/`（boundaries 规则 error 级强制，无豁免）。
- features 禁止 import `app/`、`shell/`；renderer 禁止 import `electron/*`。
- 新路由在 `src/app/router/index.ts` 登记，懒加载 `() => import('@/features/...')`，禁止静态 import 页面组件。
- 禁止新建已废除的 tech-split 目录（`@/apis`、`@/types`、`@/stores`、`@/pages`、`@/platform`、`@/layout`、`@/composables`、`@/constants`、`@/config`）。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| `.vue` 组件文件用 PascalCase | ~~todo/flash-note/schedule 26 个~~ **已全部整改（2026-08-30）** | 历史反例存档；新文件出现即打回 |
| 客户端状态目录用 `stores/` | ~~image-gen 与 settings 五域~~ **已全部整改（2026-08-30，settings 五域纯查询 store 删除后目录消失）** | 历史反例存档 |
| 子域只有页面却单开目录 | — | 开子域必须三要素齐备（契约+页面+状态），否则打平到 feature 根 |

## 自查清单

- [ ] 新 feature 从形态 A 起步，路由只指 `pages/index.vue`？
- [ ] API/类型成对落在 feature 根（`<域>.api.ts` / `<域>.type.ts`）？
- [ ] 所有新文件 kebab-case（含 .vue）？
- [ ] 没有 barrel、没有跨 feature import、没有 app/shell/electron 直引？
- [ ] 纯函数逻辑放 utils/ 且配了同名 spec？

## 范本文件

- 小而全：`src/features/image-gen/`（整目录）
- 全形态：`src/features/chat/`（整目录）
- 子域拆分：`src/features/settings/mcp/`、`src/features/todo/schedule/`

相关技能：`api-query`、`type-dto`、`pinia-store`、`code-style-guide`。
