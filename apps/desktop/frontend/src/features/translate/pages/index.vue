<template>
  <main class="mx-auto flex h-full w-full max-w-3xl flex-col gap-4 px-6 pb-6 pt-1">
    <!-- 输入区 -->
    <section class="flex flex-col gap-2">
      <Textarea
        v-model="store.input"
        autofocus
        placeholder="输入要翻译的文本，Enter 翻译，Shift+Enter 换行"
        class="min-h-20"
        @keydown="onKeydown"
      />
      <div class="flex items-center gap-2">
        <Select v-model="store.sourceLang">
          <SelectTrigger size="sm" class="w-32" aria-label="源语言">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="l in LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
          </SelectContent>
        </Select>
        <ArrowRight :size="16" class="shrink-0 text-muted-foreground" aria-hidden="true" />
        <Select v-model="store.targetLang">
          <SelectTrigger size="sm" class="w-32" aria-label="目标语言">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="l in TARGET_LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
          </SelectContent>
        </Select>
        <!-- 对照渠道多选（勾选后菜单保持打开，便于连续调整） -->
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="outline" size="sm" aria-label="选择对照渠道">
              <Layers :size="14" />
              渠道 {{ store.activeServices.length }}/{{ store.enabledServices.length }}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" class="w-56">
            <DropdownMenuLabel>对照渠道</DropdownMenuLabel>
            <DropdownMenuCheckboxItem
              v-for="s in store.enabledServices"
              :key="s.id"
              :model-value="store.selectedIds?.includes(s.id) ?? true"
              @select.prevent
              @update:model-value="(v: boolean) => store.toggleSelected(s.id, v)"
            >
              {{ s.name }}
            </DropdownMenuCheckboxItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <div class="flex-1" />
        <Button variant="ghost" :disabled="store.isTranslating" @click="store.reset()">清空</Button>
        <Button v-if="store.isTranslating" variant="outline" @click="store.cancel()">取消</Button>
        <Button v-else :disabled="!canTranslate" @click="store.translate()">翻译</Button>
      </div>
    </section>

    <!-- 结果区（多开对照）：默认隐藏，有翻译活动或无可用服务（引导卡）时展现 -->
    <section v-if="store.activeServices.length === 0 || store.hasVisibleStreams" class="flex min-h-0 flex-1 flex-col">
      <Card v-if="store.activeServices.length === 0">
        <CardContent class="flex flex-col items-center gap-1 py-10 text-center">
          <p class="text-sm font-medium">还没有启用的翻译服务</p>
          <p class="text-xs text-muted-foreground">
            前往
            <RouterLink
              to="/services"
              class="text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
            >
              服务页
            </RouterLink>
            添加并启用服务后开始翻译。
          </p>
        </CardContent>
      </Card>

      <!-- 单服务单卡片；多服务 Tabs 对照（标签 = 服务名） -->
      <stream-result v-else-if="store.activeServices.length === 1" :service-id="store.activeServices[0]!.id" />

      <Tabs
        v-else
        :key="tabKey"
        :default-value="store.activeServices[0]?.id"
        class="flex min-h-0 flex-1 flex-col gap-3"
      >
        <TabsList class="w-fit">
          <TabsTrigger v-for="s in store.activeServices" :key="s.id" :value="s.id">{{ s.name }}</TabsTrigger>
        </TabsList>
        <TabsContent v-for="s in store.activeServices" :key="s.id" :value="s.id" class="min-h-0 flex-1">
          <stream-result :service-id="s.id" />
        </TabsContent>
      </Tabs>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ArrowRight, Layers } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { LANGUAGES } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';
import { useTranslateStore } from '@/features/translate/store';
import StreamResult from '../components/stream-result.vue';

/** 目标语言选项（auto 仅用于源语言）. */
const TARGET_LANGUAGES = LANGUAGES.filter((l) => l.code !== 'auto');

const store = useTranslateStore();
const route = useRoute();
const router = useRouter();

/** 输入非空才可翻译. */
const canTranslate = computed(() => store.input.trim().length > 0);

/** 多开服务集合变化时重建 Tabs（default-value 随首个服务刷新）. */
const tabKey = computed(() => store.activeServices.map((s) => s.id).join(','));

onMounted(() => {
  void store.loadServices();
  void applyRouteText();
  void applySettingsDefaults();
});

/** Enter 触发翻译（IME 组合中不触发，Shift+Enter 换行）. */
function onKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return;
  e.preventDefault();
  if (canTranslate.value && !store.isTranslating) void store.translate();
}

/** 历史页「重新翻译」经路由 query 带入文本（features 互禁 import 的约定通道）. */
async function applyRouteText(): Promise<void> {
  const text = route.query.text;
  if (typeof text === 'string' && text) {
    store.input = text;
    await router.replace({ query: {} });
  }
}

/** 默认语言对读自设置（B6 写回 settings 的默认语言）. */
async function applySettingsDefaults(): Promise<void> {
  const bridge = await desktop();
  const s = await bridge.getSettings();
  store.sourceLang = s.sourceLang;
  store.targetLang = s.targetLang;
}
</script>
