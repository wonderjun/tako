<template>
  <main class="mx-auto flex w-full max-w-3xl flex-col gap-4 p-6">
    <header>
      <h1 class="text-lg font-semibold">设置</h1>
      <p class="text-body text-muted-foreground">语言、快捷键、外观与授权信息。</p>
    </header>

    <!-- 通用：默认语言 -->
    <Card>
      <CardContent class="grid gap-4">
        <div>
          <p class="text-sm font-semibold">通用</p>
          <p class="text-xs text-muted-foreground">翻译页的默认语言对</p>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-sm font-medium">默认源语言</p>
            <p class="text-xs text-muted-foreground">可选择自动检测。</p>
          </div>
          <Select v-model="draft.sourceLang">
            <SelectTrigger size="sm" class="w-36" aria-label="默认源语言">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="l in LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-sm font-medium">默认目标语言</p>
            <p class="text-xs text-muted-foreground">划词与输入翻译的默认结果语言。</p>
          </div>
          <Select v-model="draft.targetLang">
            <SelectTrigger size="sm" class="w-36" aria-label="默认目标语言">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="l in TARGET_LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- 快捷键：展示 + 录制 -->
    <Card>
      <CardContent class="grid gap-4">
        <div>
          <p class="text-sm font-semibold">快捷键</p>
          <p class="text-xs text-muted-foreground">点击录制后按下组合键，Esc 取消</p>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-sm font-medium">划词翻译</p>
            <p class="text-xs text-muted-foreground">选中任意文本后唤起浮窗翻译。</p>
          </div>
          <div class="flex items-center gap-2">
            <Kbd>{{ draft.hotkeys.selectionTranslate }}</Kbd>
            <Button
              size="sm"
              variant="outline"
              :disabled="recording !== null"
              @click="startRecording('selectionTranslate')"
            >
              {{ recording === 'selectionTranslate' ? '按下快捷键…' : '录制' }}
            </Button>
          </div>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-sm font-medium">输入翻译</p>
            <p class="text-xs text-muted-foreground">唤起输入窗直接打字翻译。</p>
          </div>
          <div class="flex items-center gap-2">
            <Kbd>{{ draft.hotkeys.inputTranslate }}</Kbd>
            <Button
              size="sm"
              variant="outline"
              :disabled="recording !== null"
              @click="startRecording('inputTranslate')"
            >
              {{ recording === 'inputTranslate' ? '按下快捷键…' : '录制' }}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 外观：主题 -->
    <Card>
      <CardContent class="grid gap-4">
        <div>
          <p class="text-sm font-semibold">外观</p>
          <p class="text-xs text-muted-foreground">界面明暗跟随你的选择或系统</p>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-sm font-medium">主题</p>
            <p class="text-xs text-muted-foreground">「跟随系统」时随系统深色偏好自动切换。</p>
          </div>
          <Select v-model="draft.theme">
            <SelectTrigger size="sm" class="w-36" aria-label="主题">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="light">浅色</SelectItem>
              <SelectItem value="dark">深色</SelectItem>
              <SelectItem value="system">跟随系统</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- 关于：版本与授权 -->
    <Card>
      <CardContent class="grid gap-4">
        <div>
          <p class="text-sm font-semibold">关于</p>
          <p class="text-xs text-muted-foreground">版本与授权信息</p>
        </div>
        <div class="flex items-center justify-between gap-4">
          <p class="text-sm font-medium">版本</p>
          <span class="text-sm text-muted-foreground">v{{ APP_VERSION }}</span>
        </div>
        <div class="flex items-center justify-between gap-4">
          <p class="text-sm font-medium">授权</p>
          <span class="flex items-center gap-2 text-sm text-muted-foreground">
            {{ PLAN_LABEL[license?.plan ?? 'free'] }}
            <Badge v-if="license?.licensed" variant="success">已授权</Badge>
            <Badge v-else variant="warning">未授权</Badge>
          </span>
        </div>
      </CardContent>
    </Card>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue';
import { useEventListener } from '@vueuse/core';
import { toast } from 'vue-sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Kbd } from '@/components/ui/kbd';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { LANGUAGES } from '@contracts/index';
import type { AppSettings, LicenseInfo } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';
import { useSettingsStore } from '@/features/settings/store';
import { formatHotkey } from './utils/format-hotkey';

/** 应用版本（vite define 管道未接，SPEC-B 允许硬编码）. */
const APP_VERSION = '0.1.0';

/** 授权档位文案映射. */
const PLAN_LABEL: Record<LicenseInfo['plan'], string> = {
  free: '免费版',
  pro: '专业版',
  trial: '试用版',
};

/** 目标语言选项（auto 仅用于源语言）. */
const TARGET_LANGUAGES = LANGUAGES.filter((l) => l.code !== 'auto');

const settingsStore = useSettingsStore();

/** 设置草稿：控件写草稿 → 即时 persist 写回 store/bridge. */
const draft = reactive<AppSettings>({
  sourceLang: 'auto',
  targetLang: 'zh-Hans',
  hotkeys: { selectionTranslate: 'Alt+D', inputTranslate: 'Alt+A' },
  theme: 'system',
  launchAtLogin: false,
  maxConcurrent: 5,
});

/** 授权信息. */
const license = ref<LicenseInfo | null>(null);

/** 正在录制的快捷键字段. */
const recording = ref<null | 'selectionTranslate' | 'inputTranslate'>(null);

// 任一设置变更即写回（主题切换即时生效依赖 settings store 的副作用）
watch(
  draft,
  () => {
    void persist();
  },
  { deep: true },
);

onMounted(async () => {
  try {
    await settingsStore.load();
    if (settingsStore.settings) Object.assign(draft, settingsStore.settings);
    const bridge = await desktop();
    license.value = await bridge.licenseInfo();
  } catch {
    toast.error('设置加载失败');
  }
});

/** 草稿写回（主题即时生效由 settings store 的副作用完成）. */
async function persist(): Promise<void> {
  try {
    await settingsStore.save({ ...draft, hotkeys: { ...draft.hotkeys } });
  } catch {
    toast.error('设置保存失败');
  }
}

function startRecording(field: 'selectionTranslate' | 'inputTranslate'): void {
  recording.value = field;
}

/**
 * 录制捕获（window capture）：Esc 取消，纯修饰键忽略，有效组合写回.
 * 挂 window 而非 <main>——录制按钮 disabled 后焦点可能落在 body，
 * target=body 的 keydown 不经过 main 子树（曾致"按键没反应"）.
 */
useEventListener(
  window,
  'keydown',
  (e: KeyboardEvent) => {
    if (recording.value === null) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.key === 'Escape') {
      recording.value = null;
      return;
    }
    const hotkey = formatHotkey(e);
    if (!hotkey) return;
    draft.hotkeys[recording.value] = hotkey;
    recording.value = null;
  },
  { capture: true },
);
</script>
