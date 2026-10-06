# ==============================================================================
# Tako 全仓构建入口（Wails 3 一体化：Go 壳 + Vue 前端）
# ==============================================================================
# 流程约定：
#   make dev     开发模式：wails3 dev（vite dev server + go run 热重载）
#   make build   生产构建：当前平台安装包（macOS → dmg；Windows → nsis）
#   make ci      本地 CI：前端 lint/typecheck/test + Go vet/test
#
# 跨平台边界：wails3 只能打宿主机平台的包，全量发布需在两个平台各跑一次 make build。
# ==============================================================================

DESKTOP_DIR := apps/desktop
FRONTEND_DIR := apps/desktop/frontend

.PHONY: dev build ci generate-bindings ci-frontend ci-frontend-install ci-go generate clean

# ------------------------------------------------------------------------------
# dev: 开发模式（热重载）
# ------------------------------------------------------------------------------
dev:
	cd $(DESKTOP_DIR) && wails3 dev

# ------------------------------------------------------------------------------
# build: 生产构建当前平台安装包
# ------------------------------------------------------------------------------
build:
	cd $(DESKTOP_DIR) && wails3 build

# ------------------------------------------------------------------------------
# generate: 重新生成 Wails bindings（bridge 方法/类型变更后执行并提交产物）
# flags 与 wails3 build 内部一致（-ts -i = TypeScript + interfaces），
# 生成物为 .ts（带类型）且随仓库入库。
# ------------------------------------------------------------------------------
generate:
	cd $(DESKTOP_DIR) && wails3 generate bindings -ts -i -clean=true

# ------------------------------------------------------------------------------
# ci: 全部本地 CI 检查（push 前必跑）
# bindings 已入库，无需先生成
# ------------------------------------------------------------------------------
ci: ci-frontend ci-go

# ------------------------------------------------------------------------------
# ci-frontend: 前端 install + lint + 类型检查 + 测试
# ------------------------------------------------------------------------------
ci-frontend: ci-frontend-install
	cd $(FRONTEND_DIR) && bun run lint && bun run typecheck && bun run test

ci-frontend-install:
	cd $(FRONTEND_DIR) && bun install --frozen-lockfile

# ------------------------------------------------------------------------------
# ci-go: Go 静态检查 + 测试（排除 frontend/node_modules 内混入的 go 文件）
# ------------------------------------------------------------------------------
ci-go:
	cd $(DESKTOP_DIR) && go vet $$(go list ./... | grep -v node_modules) && go test $$(go list ./... | grep -v node_modules)

# ------------------------------------------------------------------------------
# clean: 清理构建产物
# ------------------------------------------------------------------------------
clean:
	rm -rf $(DESKTOP_DIR)/build/bin $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules/.vite
