<template>
  <div class="card p-3 grid grid-cols-2 items-center justify-center cursor-pointer bg-base-200 hover:bg-base-300 transition-colors" @click="$emit('click')">
    <Icon color="white" :name="icon" size="60" />
    <div class="space-y-2">
      <p class="title">{{ label }}</p>
      <Icon v-if="loadingOverview" name="mdi:loading" size="20" class="animate-spin text-base-content" />
      <div v-else :class="['w-fit text-xs px-2 py-1 rounded mb-1', installed ? 'bg-success text-success-content' : 'bg-base-300 text-base-content']">
        {{ version || "Unavailable" }}
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  languageKey: {
    type: String,
    required: true
  },
  version: {
    type: String,
    default: ""
  },
  installed: {
    type: Boolean,
    default: false
  },
  loadingOverview: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

// Language configurations
const languageConfigs = {
  php: {
    label: 'PHP',
    icon: 'vscode-icons:file-type-php'
  },
  node: {
    label: 'Node.js',
    icon: 'vscode-icons:file-type-node'
  },
  python: {
    label: 'Python',
    icon: 'vscode-icons:file-type-python'
  },
  golang: {
    label: 'Go',
    icon: 'vscode-icons:file-type-go'
  }
}

// Get label and icon based on language key
const label = computed(() => languageConfigs[props.languageKey]?.label || props.languageKey)
const icon = computed(() => languageConfigs[props.languageKey]?.icon || 'mdi:package-variant')
</script>
