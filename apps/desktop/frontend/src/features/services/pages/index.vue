<template>
  <main class="mx-auto flex w-full max-w-4xl flex-col gap-4 px-6 pb-6 pt-1">
    <!-- 页头（标题在壳层标题栏，见 app-shell + route.meta.title） -->
    <header class="flex items-center justify-between">
      <p class="text-body text-muted-foreground">管理多开对照的翻译服务，启用的服务参与翻译。</p>
      <Button @click="openCreate()">
        <Plus :size="16" />
        添加服务
      </Button>
    </header>

    <!-- 服务卡片列表 -->
    <Card v-for="s in servicesStore.services" :key="s.id">
      <CardContent class="flex items-center gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="truncate text-sm font-medium">{{ s.name }}</span>
            <Badge variant="secondary">{{ TYPE_LABEL[s.type] }}</Badge>
          </div>
          <p class="mt-0.5 truncate text-xs text-muted-foreground">{{ summaryOf(s) }}</p>
        </div>
        <Switch :model-value="s.enabled" aria-label="启用服务" @update:model-value="() => onToggle(s)" />
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="icon-sm" aria-label="更多操作">
              <EllipsisVertical :size="16" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem @click="openEdit(s)">
              <Pencil :size="16" />
              编辑
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" @click="onDelete(s)">
              <Trash2 :size="16" />
              删除
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardContent>
    </Card>

    <!-- 空态 -->
    <Card v-if="!servicesStore.services.length && !servicesStore.isLoading">
      <CardContent class="flex flex-col items-center gap-1 py-8 text-center">
        <p class="text-sm font-medium">还没有翻译服务</p>
        <p class="text-xs text-muted-foreground">点击「添加服务」配置第一个服务，开始多开对照。</p>
      </CardContent>
    </Card>

    <!-- 编辑弹窗 -->
    <ServiceDialog v-model:open="dialogOpen" :service="editing" />
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { toast } from 'vue-sonner';
import { EllipsisVertical, Pencil, Plus, Trash2 } from '@lucide/vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';
import type { ProviderType, ServiceSetting } from '@contracts/index';
import { useServicesStore } from '@/features/services/store';
import ServiceDialog from './components/service-dialog.vue';

/** 服务类型文案映射. */
const TYPE_LABEL: Record<ProviderType, string> = {
  llm: 'LLM',
  deepl: 'DeepL',
  google: 'Google',
  microsoft: 'Microsoft',
};

const servicesStore = useServicesStore();

/** 弹窗开关与编辑目标. */
const dialogOpen = ref(false);
const editing = ref<ServiceSetting | null>(null);

onMounted(() => {
  void servicesStore.load();
});

/** 卡片副标题：llm 展示连接信息，传统服务展示类型说明. */
function summaryOf(s: ServiceSetting): string {
  if (s.type === 'llm') {
    return [s.baseUrl, s.model].filter(Boolean).join(' · ') || '未配置连接信息';
  }
  return `${TYPE_LABEL[s.type]} 官方接口`;
}

function openCreate(): void {
  editing.value = null;
  dialogOpen.value = true;
}

function openEdit(s: ServiceSetting): void {
  editing.value = s;
  dialogOpen.value = true;
}

/** 启用开关（失败 toast，乐观回滚在 store 内）. */
async function onToggle(s: ServiceSetting): Promise<void> {
  try {
    await servicesStore.toggle(s.id);
  } catch {
    toast.error('切换失败，请重试');
  }
}

/** 删除服务. */
async function onDelete(s: ServiceSetting): Promise<void> {
  try {
    await servicesStore.remove(s.id);
    toast.success('服务已删除');
  } catch {
    toast.error('删除失败，请重试');
  }
}
</script>
