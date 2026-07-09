<template>
  <div class="form-control w-full">
    <!-- File input variant -->
    <label v-if="variant === 'file'" class="input input-bordered flex items-center gap-2 w-full">
      <Icon v-if="icon" :name="icon" class="opacity-50" size="20" />
      <input :value="modelValue" @input="$emit('update:modelValue', $event.target.value)" type="text" class="grow w-full"
        :placeholder="placeholder" :name="label" :required="required" :disabled="disabled" :readonly="readonly" />
    </label>

    <!-- Textarea with label -->
    <div v-else-if="textarea" class="form-control w-full">
      <label v-if="label" class="label">
        <span class="label-text">
          {{ label }}
          <span v-if="required" class="text-error ml-1">*</span>
        </span>
      </label>
      <textarea :value="modelValue" @input="$emit('update:modelValue', $event.target.value)"
        class="textarea textarea-bordered h-24 w-full" :class="[
          monospace ? 'font-mono text-sm' : '',
          disabled ? 'opacity-50 cursor-not-allowed' : ''
        ]"
        :placeholder="placeholder" :name="label" :required="required" :disabled="disabled" :readonly="readonly"
        :rows="rows" />
    </div>

    <!-- Regular input with inline label -->
    <label v-else class="input input-bordered flex items-center gap-2 w-full">
      {{ label }}
      <span v-if="required" class="text-error ml-1">*</span>
      <Icon v-if="icon" :name="icon" class="opacity-50" size="20" />
      <input :value="modelValue" @input="$emit('update:modelValue', $event.target.value)" :type="type" class="grow w-full"
        :class="[
          monospace ? 'font-mono text-sm' : '',
          disabled ? 'opacity-50 cursor-not-allowed' : ''
        ]"
        :placeholder="placeholder" :name="label" :required="required" :disabled="disabled" :readonly="readonly" />
      <span v-if="badge" class="badge badge-neutral badge-xs">{{ badge }}</span>
      <button v-if="type === 'password' && appendIcon" type="button" @click="$emit('click:append')" class="btn btn-ghost btn-xs">
        <Icon :name="appendIcon" size="16" />
      </button>
    </label>

    <label v-if="hint" class="label">
      <span class="label-text-alt">{{ hint }}</span>
    </label>
  </div>
</template>

<script setup>
defineEmits(['update:modelValue', 'click:append'])

const modelValue = defineModel()

const props = defineProps({
  label: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'text'
  },
  placeholder: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
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
  readonly: {
    type: Boolean,
    default: false
  },
  monospace: {
    type: Boolean,
    default: false
  },
  hint: {
    type: String,
    default: ''
  },
  textarea: {
    type: Boolean,
    default: false
  },
  rows: {
    type: Number,
    default: 4
  },
  variant: {
    type: String,
    default: '',
    validator: (value) => ['', 'file'].includes(value)
  },
  badge: {
    type: String,
    default: ''
  },
  appendIcon: {
    type: String,
    default: ''
  }
})
</script>