<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? '编辑服务' : '添加服务' }}</DialogTitle>
        <DialogDescription>配置翻译服务的连接信息；llm 类型走 OpenAI 兼容协议并支持流式对照。</DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" novalidate @submit="submit">
        <!-- 公共字段：名称 -->
        <FormField v-slot="{ componentField }" name="name">
          <FormItem>
            <FormLabel>名称</FormLabel>
            <FormControl>
              <Input v-bind="componentField" placeholder="例如 My OpenAI" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <!-- 公共字段：类型 -->
        <FormField v-slot="{ componentField }" name="type">
          <FormItem>
            <FormLabel>类型</FormLabel>
            <Select v-bind="componentField">
              <FormControl>
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="选择类型" />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                <SelectItem v-for="t in TYPE_OPTIONS" :key="t.value" :value="t.value">{{ t.label }}</SelectItem>
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        </FormField>

        <!-- llm 专用字段（v-show 保持挂载，切换类型不丢已填值） -->
        <div v-show="values.type === 'llm'" class="grid gap-4">
          <FormField v-slot="{ componentField }" name="baseUrl">
            <FormItem>
              <FormLabel>Base URL</FormLabel>
              <FormControl>
                <Input v-bind="componentField" placeholder="如 api.example.com/v1" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="apiKey">
            <FormItem>
              <FormLabel>API Key</FormLabel>
              <FormControl>
                <Input v-bind="componentField" type="password" placeholder="sk-..." />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="model">
            <FormItem>
              <FormLabel>模型</FormLabel>
              <FormControl>
                <Input v-bind="componentField" placeholder="例如 gpt-4o-mini" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="authScheme">
            <FormItem>
              <FormLabel>鉴权方式</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger class="w-full">
                    <SelectValue placeholder="选择鉴权方式" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="bearer">Bearer Token（Authorization 头）</SelectItem>
                  <SelectItem value="x-api-key">x-api-key 头</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>

        <!-- 连通性测试结果 -->
        <Alert v-if="testResult?.ok" variant="success">
          <Check />
          <AlertTitle>连通正常 · {{ testResult.latencyMs }}ms</AlertTitle>
          <AlertDescription>样例译文：{{ testResult.sample }}</AlertDescription>
        </Alert>
        <Alert v-else-if="testResult" variant="destructive">
          <CircleAlert />
          <AlertTitle>测试失败</AlertTitle>
          <AlertDescription>{{ testResult.error ?? '未知错误' }}</AlertDescription>
        </Alert>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="isTesting" @click="runTest()">
            <Spinner v-if="isTesting" class="size-3.5" />
            测试
          </Button>
          <Button type="submit" :disabled="isSubmitting">保存</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import { z } from 'zod';
import { toast } from 'vue-sonner';
import { Check, CircleAlert } from '@lucide/vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import type { ProviderType, ServiceSetting, TestResult } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';
import { useServicesStore, type ServiceDraft } from '@/features/services/store';

const props = defineProps<{
  /** 弹窗是否打开. */
  open: boolean;
  /** 编辑目标；null 表示新增. */
  service: ServiceSetting | null;
}>();

const emit = defineEmits<{
  'update:open': [open: boolean];
}>();

/** 类型选项文案映射. */
const TYPE_OPTIONS: { value: ProviderType; label: string }[] = [
  { value: 'llm', label: 'LLM（OpenAI 兼容）' },
  { value: 'deepl', label: 'DeepL' },
  { value: 'google', label: 'Google 翻译' },
  { value: 'microsoft', label: 'Microsoft 翻译' },
];

/** 表单校验：name 必填；llm 时 baseUrl 必填. */
const schema = z
  .object({
    name: z.string().min(1, '请输入服务名称'),
    type: z.enum(['llm', 'deepl', 'google', 'microsoft']),
    baseUrl: z.string(),
    apiKey: z.string(),
    model: z.string(),
    authScheme: z.enum(['bearer', 'x-api-key']),
  })
  .superRefine((val, ctx) => {
    if (val.type === 'llm' && !val.baseUrl.trim()) {
      ctx.addIssue({ code: 'custom', message: 'llm 服务需填写 Base URL', path: ['baseUrl'] });
    }
  });

type ServiceFormValues = z.infer<typeof schema>;

/** 是否编辑态. */
const isEdit = computed(() => props.service !== null);

/** 表单初值（新增时给 llm 默认项）. */
const initialValues = computed(() => ({
  name: props.service?.name ?? '',
  type: props.service?.type ?? ('llm' as ProviderType),
  baseUrl: props.service?.baseUrl ?? '',
  apiKey: props.service?.apiKey ?? '',
  model: props.service?.model ?? '',
  authScheme: props.service?.authScheme ?? ('bearer' as 'bearer' | 'x-api-key'),
}));

const { handleSubmit, values, isSubmitting, resetForm } = useForm({
  // 裸 zod schema 不被 vee-validate 识别（校验静默失效），必须经 toTypedSchema 适配
  validationSchema: toTypedSchema(schema),
  initialValues: initialValues.value,
});

/** 连通性测试状态. */
const isTesting = ref(false);
const testResult = ref<TestResult | null>(null);

// 每次打开按当前编辑目标重置表单与测试结果
watch(
  () => props.open,
  (open) => {
    testResult.value = null;
    if (open) resetForm({ values: initialValues.value });
  },
);

/** 表单值 → 编辑草稿（非 llm 类型剥离 llm 专用字段）. */
function buildDraft(values: ServiceFormValues): ServiceDraft {
  const base = {
    id: props.service?.id ?? '',
    name: values.name.trim(),
    type: values.type,
    enabled: props.service?.enabled ?? true,
    order: props.service?.order,
  };
  if (values.type !== 'llm') return base;
  return {
    ...base,
    baseUrl: values.baseUrl.trim(),
    apiKey: values.apiKey.trim() || undefined,
    model: values.model.trim() || undefined,
    authScheme: values.authScheme,
  };
}

/** 表单值 → 可测试的服务配置（契约允许未保存配置直接测试）. */
function buildTestService(values: ServiceFormValues): ServiceSetting {
  return { ...buildDraft(values), order: 0, createdAt: 0, updatedAt: 0 };
}

/** 保存（成功关弹窗，失败 toast 保留弹窗重试）. */
const submit = handleSubmit(async (formValues) => {
  const servicesStore = useServicesStore();
  try {
    await servicesStore.save(buildDraft(formValues as ServiceFormValues));
    toast.success(isEdit.value ? '服务已保存' : '服务已添加');
    emit('update:open', false);
  } catch {
    toast.error('保存失败，请重试');
  }
});

/** 连通性测试（loading → 成功显示样例与延迟 / 失败显示 error）. */
async function runTest(): Promise<void> {
  isTesting.value = true;
  testResult.value = null;
  try {
    const bridge = await desktop();
    testResult.value = await bridge.testService(buildTestService(values as ServiceFormValues));
  } catch {
    testResult.value = { ok: false, latencyMs: 0, error: '测试请求失败' };
  } finally {
    isTesting.value = false;
  }
}
</script>
