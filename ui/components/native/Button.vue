<template>
  <button 
    :type="type"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
    :class="[
      'flex items-center justify-center gap-2 px-6 py-3 rounded-lg font-medium transition-colors',
      buttonClasses,
      disabled || loading ? 'cursor-not-allowed' : 'cursor-pointer',
      fullWidth ? 'w-full' : '',
      size === 'sm' ? 'px-4 py-2 text-sm' : '',
      size === 'lg' ? 'px-8 py-4 text-lg' : ''
    ]"
  >
    <Icon 
      v-if="loading" 
      name="mdi:loading" 
      class="animate-spin" 
      :size="iconSize" 
    />
    <Icon 
      v-else-if="icon" 
      :name="icon" 
      :size="iconSize" 
    />
    <slot />
  </button>
</template>

<script setup>
defineEmits(['click'])

const props = defineProps({
  type: {
    type: String,
    default: 'button'
  },
  variant: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'secondary', 'success', 'danger', 'warning'].includes(value)
  },
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['sm', 'md', 'lg'].includes(value)
  },
  disabled: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  icon: {
    type: String,
    default: ''
  },
  fullWidth: {
    type: Boolean,
    default: false
  }
})

const buttonClasses = computed(() => {
  const variants = {
    primary: 'btn btn-primary',
    secondary: 'btn btn-secondary',
    success: 'btn btn-success',
    danger: 'btn btn-error',
    warning: 'btn btn-warning'
  }
  return variants[props.variant] || variants.primary
})

const iconSize = computed(() => {
  const sizes = {
    sm: '16',
    md: '20',
    lg: '24'
  }
  return sizes[props.size] || sizes.md
})
</script>