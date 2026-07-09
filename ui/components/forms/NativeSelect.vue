<template>
  <div class="form-control w-full">
    <label class="label">
      <span class="label-text font-medium">{{ label }}</span>
      <span v-if="required" class="label-text-alt text-error">*</span>
    </label>
    <select
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      :required="required"
      :disabled="disabled"
      class="select select-bordered select-primary w-full"
      :class="{ 'select-error': error }"
    >
      <option value="" disabled>{{ placeholder || 'Select an option' }}</option>
      <option
        v-for="option in options"
        :key="option"
        :value="option"
        :selected="option === modelValue || option === default"
      >
        {{ option }}
      </option>
    </select>
    <label v-if="helper" class="label">
      <span class="label-text-alt text-base-content/60">{{ helper }}</span>
    </label>
    <label v-if="error" class="label">
      <span class="label-text-alt text-error">{{ error }}</span>
    </label>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  label: {
    type: String,
    required: true
  },
  options: {
    type: Array,
    required: true
  },
  placeholder: {
    type: String,
    default: ''
  },
  default: {
    type: [String, Number],
    default: ''
  },
  required: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  helper: {
    type: String,
    default: ''
  }
})

defineEmits(['update:modelValue'])
</script>
