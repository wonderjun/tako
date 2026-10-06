<template>
  <main class="mx-auto flex h-full w-full max-w-4xl flex-col gap-4 px-6 pb-6 pt-1">
    <!-- 页头（标题在壳层标题栏，见 app-shell + route.meta.title） -->
    <header class="flex items-center justify-between">
      <p class="text-body text-muted-foreground">共 {{ store.total }} 条翻译记录。</p>
    </header>

    <!-- 过滤条 -->
    <div class="flex items-center gap-2">
      <div class="relative">
        <Search
          :size="14"
          class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        <Input
          v-model="keywordInput"
          placeholder="搜索原文或译文"
          size="sm"
          class="w-56 pl-7"
          aria-label="关键词搜索"
        />
      </div>
      <Select v-model="serviceFilter">
        <SelectTrigger size="sm" class="w-40" aria-label="按服务过滤">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">全部服务</SelectItem>
          <SelectItem v-for="s in serviceOptions" :key="s.id" :value="s.id">{{ s.name }}</SelectItem>
        </SelectContent>
      </Select>
      <label class="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Switch
          :model-value="store.query.favoriteOnly ?? false"
          @update:model-value="(v: boolean) => (store.query.favoriteOnly = v)"
        />
        仅收藏
      </label>
      <div class="flex-1" />
      <AlertDialog v-model:open="clearOpen">
        <AlertDialogTrigger as-child>
          <Button variant="destructive" size="sm">清空历史</Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>清空全部历史？</AlertDialogTitle>
            <AlertDialogDescription>将删除全部翻译历史与收藏记录，此操作不可撤销。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <!-- 破坏性异步操作用普通 Button 触发，成功后再关弹窗（失败保留重试） -->
            <Button variant="destructive" :disabled="store.isLoading" @click="onClearAll">确认清空</Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>

    <!-- 列表（内部滚动容器，触底加载下一页） -->
    <div ref="scrollRef" class="min-h-0 flex-1 overflow-y-auto rounded-md border border-hairline">
      <Table>
        <TableHeader class="sticky top-0 bg-card">
          <TableRow>
            <TableHead class="w-24">时间</TableHead>
            <TableHead>原文</TableHead>
            <TableHead>译文</TableHead>
            <TableHead class="w-28">服务</TableHead>
            <TableHead class="w-14">收藏</TableHead>
            <TableHead class="w-28 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="item in store.items" :key="item.id">
            <TableCell class="whitespace-nowrap text-xs text-muted-foreground">
              {{ formatRelative(item.createdAt) }}
            </TableCell>
            <TableCell class="max-w-52">
              <span class="block truncate" :title="item.sourceText">{{ truncate(item.sourceText) }}</span>
            </TableCell>
            <TableCell class="max-w-52">
              <span class="block truncate" :title="item.resultText">{{ truncate(item.resultText) }}</span>
            </TableCell>
            <TableCell>
              <Badge variant="secondary">{{ item.serviceName }}</Badge>
            </TableCell>
            <TableCell>
              <Button
                variant="ghost"
                size="icon-xs"
                :aria-label="item.favorite ? '取消收藏' : '收藏'"
                :title="item.favorite ? '取消收藏' : '收藏'"
                @click="onToggleFavorite(item)"
              >
                <Star :size="14" :class="item.favorite ? 'fill-current text-warning' : 'text-muted-foreground'" />
              </Button>
            </TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="icon-xs" aria-label="查看详情" title="查看详情" @click="detailItem = item">
                <Eye :size="14" />
              </Button>
              <Button
                variant="ghost"
                size="icon-xs"
                aria-label="重新翻译"
                title="重新翻译"
                @click="onRetranslate(item)"
              >
                <RotateCcw :size="14" />
              </Button>
              <Button variant="ghost-destructive" size="icon-xs" aria-label="删除" title="删除" @click="onDelete(item)">
                <Trash2 :size="14" />
              </Button>
            </TableCell>
          </TableRow>

          <!-- 空态与加载态 -->
          <TableRow v-if="store.items.length === 0 && !store.isLoading">
            <TableCell :colspan="6" class="h-24 text-center text-sm text-muted-foreground">暂无历史记录</TableCell>
          </TableRow>
          <TableRow v-if="store.isLoading && store.items.length === 0">
            <TableCell :colspan="6" class="h-24 text-center text-sm text-muted-foreground">加载中…</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
    <p v-if="store.isLoading && store.items.length > 0" class="text-center text-xs text-muted-foreground">加载中…</p>

    <!-- 详情弹窗（全文） -->
    <Dialog
      :open="detailItem !== null"
      @update:open="
        (v: boolean) => {
          if (!v) detailItem = null;
        }
      "
    >
      <DialogContent class="max-w-xl">
        <DialogHeader>
          <DialogTitle>翻译详情</DialogTitle>
          <DialogDescription v-if="detailItem">
            {{ detailItem.serviceName }} · {{ languageLabel(detailItem.sourceLang) }} →
            {{ languageLabel(detailItem.targetLang) }} ·
            {{ formatRelative(detailItem.createdAt) }}
          </DialogDescription>
        </DialogHeader>
        <div v-if="detailItem" class="grid gap-3">
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">原文</p>
            <p class="whitespace-pre-wrap rounded-md border border-hairline bg-muted/40 p-3 text-body">
              {{ detailItem.sourceText }}
            </p>
          </div>
          <div>
            <p class="mb-1 text-xs font-medium text-muted-foreground">译文</p>
            <p class="whitespace-pre-wrap rounded-md border border-hairline bg-muted/40 p-3 text-body">
              {{ detailItem.resultText }}
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useInfiniteScroll, refDebounced } from '@vueuse/core';
import { Eye, RotateCcw, Search, Star, Trash2 } from '@lucide/vue';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from 'vue-sonner';
import type { HistoryItem, ServiceSetting } from '@contracts/index';
import { languageLabel } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';
import { useHistoryStore } from '@/features/history/store';
import { formatRelative } from './utils/format-relative';

const store = useHistoryStore();
const router = useRouter();

/** 关键词输入（300ms 防抖后生效）. */
const keywordInput = ref('');
const keyword = refDebounced(keywordInput, 300);

/** 服务过滤（'all' 为哨兵，reka Select 不接受空串 value）. */
const serviceFilter = ref('all');

/** 触底加载的滚动容器. */
const scrollRef = ref<HTMLElement | null>(null);

/** 服务过滤选项快照. */
const serviceOptions = ref<ServiceSetting[]>([]);

/** 详情弹窗目标. */
const detailItem = ref<HistoryItem | null>(null);

/** 清空确认弹窗. */
const clearOpen = ref(false);

useInfiniteScroll(
  scrollRef,
  () => {
    void store.loadMore();
  },
  { distance: 12 },
);

watch([keyword, () => store.query.favoriteOnly], () => {
  void store.load();
});

watch(serviceFilter, (v) => {
  store.query.serviceId = v === 'all' ? undefined : v;
  void store.load();
});

onMounted(async () => {
  await store.load();
  const bridge = await desktop();
  serviceOptions.value = await bridge.listServices();
});

/** 截断展示文本（60 字）. */
function truncate(text: string, max = 60): string {
  return text.length > max ? `${text.slice(0, max)}…` : text;
}

/** 收藏切换（失败 toast，乐观回滚在 store 内）. */
async function onToggleFavorite(item: HistoryItem): Promise<void> {
  try {
    await store.toggleFavorite(item);
  } catch {
    toast.error('操作失败，请重试');
  }
}

/** 重新翻译：跳翻译页并经路由 query 带入原文. */
function onRetranslate(item: HistoryItem): void {
  void router.push({ path: '/', query: { text: item.sourceText } });
}

/** 删除单条. */
async function onDelete(item: HistoryItem): Promise<void> {
  try {
    await store.remove(item.id);
    toast.success('已删除');
  } catch {
    toast.error('删除失败，请重试');
  }
}

/** 清空全部（成功后关确认弹窗）. */
async function onClearAll(): Promise<void> {
  try {
    await store.clearAll();
    clearOpen.value = false;
    toast.success('历史已清空');
  } catch {
    toast.error('清空失败，请重试');
  }
}
</script>
