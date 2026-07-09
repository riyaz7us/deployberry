<template>
  <div class="bg-base-200 border border-base-300 rounded-lg p-4 transition-all duration-200 hover:bg-base-300">
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2">
        <Icon :name="icon" size="20" class="text-primary/80" />
        <h3 class="font-medium text-base-content capitalize">{{ name }}</h3>
      </div>
      <div class="flex items-center">
        <Icon name="mdi:loading" class="animate-spin text-primary" v-if="loading" size="20" />
        <Icon name="mdi:check-circle" class="text-success" v-else-if="available" size="20" />
        <Icon name="mdi:alert-circle" class="text-error" v-else size="20" />
      </div>
    </div>
    <div class="text-sm">
      <span :class="available ? 'text-success/80' : 'text-error/80'">
        {{ version || "Not installed" }}
      </span>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  name: {
    type: String,
    required: true
  },
  icon: {
    type: String,
    required: true
  },
  available: {
    type: Boolean,
    default: false
  },
  version: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  }
})
</script>
