<template>
  <div class="space-y-6">
    <!-- Tab Navigation -->
    <div class="border-b border-base-300">
      <nav class="flex space-x-8" aria-label="Tabs">
        <button
          v-for="(tab, index) in tabs"
          :key="index"
          @click="activeTab = index"
          :class="[
            'whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors',
            activeTab === index
              ? 'border-primary text-primary'
              : 'border-transparent text-base-content/70 hover:text-base-content hover:border-base-300'
          ]"
          :aria-current="activeTab === index ? 'page' : undefined"
        >
          <div class="flex items-center gap-2">
            <Icon
              v-if="tab.icon"
              :name="tab.icon"
              :size="16"
            />
            <span>{{ tab.title }}</span>
          </div>
        </button>
      </nav>
    </div>

    <!-- Tab Content -->
    <div class="min-h-[400px]">
      <slot :name="`tab-${activeTab}`">
        <div class="text-center py-8 text-base-content/60">
          <Icon name="mdi:file-document-outline" size="48" class="mx-auto mb-4 opacity-50" />
          <p>Content for {{ tabs[activeTab]?.title || 'this tab' }} will be displayed here</p>
        </div>
      </slot>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  tabs: {
    type: Array,
    required: true,
    validator: (value) => {
      return value.every(tab =>
        typeof tab === 'object' &&
        typeof tab.title === 'string' &&
        (tab.icon === undefined || typeof tab.icon === 'string')
      )
    }
  },
  defaultActive: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['tab-changed'])

const activeTab = ref(props.defaultActive)

const activeTabData = computed(() => props.tabs[activeTab.value])

// Watch for tab changes and emit event
watch(activeTab, (newTab, oldTab) => {
  if (newTab !== oldTab) {
    emit('tab-changed', { index: newTab, tab: props.tabs[newTab] })
  }
})

// Expose methods for parent components
defineExpose({
  setActiveTab: (index) => {
    if (index >= 0 && index < props.tabs.length) {
      activeTab.value = index
    }
  },
  getActiveTab: () => activeTab.value,
  getActiveTabData: () => activeTabData.value
})
</script>
