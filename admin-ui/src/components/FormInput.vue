<template>
  <input
    :type="type"
    :value="modelValue"
    @input="$emit('update:modelValue', $event.target.value)"
    :placeholder="placeholder"
    :required="required"
    :readonly="readonly"
    class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
    :class="[mono ? 'font-mono text-xs' : '', inputClass]"
  />
</template>

<script setup>
defineProps({
  modelValue: { type: [String, Number], default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  required: { type: Boolean, default: false },
  readonly: { type: Boolean, default: false },
  mono: { type: Boolean, default: false },
  inputClass: { type: String, default: '' },
})

defineEmits(['update:modelValue'])
</script>
