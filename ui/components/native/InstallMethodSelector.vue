<template>
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div
      v-for="option in options"
      :key="option.value"
      @click="$emit('update:modelValue', option.value)"
      :class="[
        'bg-base-200 border rounded-lg p-4 cursor-pointer transition-all duration-200',
        modelValue === option.value
          ? 'border-primary bg-primary/10'
          : 'border-base-300 hover:border-base-400'
      ]"
    >
      <div class="flex items-center gap-3 mb-3">
        <Icon 
          :name="option.icon || (modelValue === option.value ? 'mdi:circle' : 'mdi:circle-outline')" 
          size="24" 
          :class="option.iconColor || (modelValue === option.value ? 'text-primary' : 'text-base-content/60')" 
        />
        <span class="text-base-content font-medium">{{ option.label }}</span>
      </div>
      <p class="text-base-content/60 text-sm">{{ option.description }}</p>
    </div>
  </div>
</template>

<script setup>
defineProps({
  options: {
    type: Array,
    required: true,
    default: () => []
  },
  modelValue: {
    type: String,
    required: true
  }
})

defineEmits(['update:modelValue'])
</script>
