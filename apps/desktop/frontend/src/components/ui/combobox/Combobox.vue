<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { computed, ref, watch } from 'vue';
import { Check, ChevronsUpDown } from '@lucide/vue';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';

export interface IComboboxOption {
  label: string;
  value: string;
  disabled?: boolean;
}

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    options: IComboboxOption[];
    placeholder?: string;
    emptyText?: string;
    disabled?: boolean;
    class?: HTMLAttributes['class'];
    triggerClass?: HTMLAttributes['class'];
    searchPlaceholder?: string;
  }>(),
  {
    modelValue: '',
    placeholder: '请选择...',
    emptyText: '未找到匹配项',
    searchPlaceholder: '搜索...',
    disabled: false,
  },
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const open = ref(false);

const selectedLabel = computed(() => {
  const found = props.options.find((o) => o.value === props.modelValue);
  return found?.label ?? '';
});

const handleSelect = (value: string) => {
  emit('update:modelValue', value);
  open.value = false;
};

// modelValue 改变时不需要任何额外处理
watch(
  () => props.modelValue,
  () => {
    /* keep open state independent */
  },
);
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <!-- CUSTOMIZED: 补 size="lg" — Button v2 默认档 36→32px, combobox 触发器保持原 36px -->
      <Button
        type="button"
        variant="outline"
        size="lg"
        role="combobox"
        :aria-expanded="open"
        :disabled="disabled"
        :class="cn('w-full justify-between font-normal', props.class, triggerClass)"
      >
        <span :class="cn('truncate', !selectedLabel && 'text-muted-foreground')">
          {{ selectedLabel || placeholder }}
        </span>
        <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-[var(--reka-popover-trigger-width)] p-0" align="start">
      <Command>
        <CommandInput :placeholder="searchPlaceholder" />
        <CommandList>
          <CommandEmpty>{{ emptyText }}</CommandEmpty>
          <CommandGroup>
            <CommandItem
              v-for="option in options"
              :key="option.value"
              :value="option.value"
              :disabled="option.disabled"
              @select="() => handleSelect(option.value)"
            >
              <Check :class="cn('mr-2 size-4', modelValue === option.value ? 'opacity-100' : 'opacity-0')" />
              <span class="truncate">{{ option.label }}</span>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
