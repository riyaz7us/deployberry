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
  databaseKey: {
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

// Database configurations
const databaseConfigs = {
  mysql: {
    label: 'MySQL',
    icon: 'mdi:database'
  },
  mariadb: {
    label: 'MariaDB',
    icon: 'mdi:database-outline'
  },
  postgres: {
    label: 'PostgreSQL',
    icon: 'mdi:database'
  },
  mongodb: {
    label: 'MongoDB',
    icon: 'mdi:database'
  },
  redis: {
    label: 'Redis',
    icon: 'mdi:database'
  },
  sqlite: {
    label: 'SQLite',
    icon: 'mdi:database'
  }
}

// Get label and icon based on database key
const label = computed(() => databaseConfigs[props.databaseKey]?.label || props.databaseKey)
const icon = computed(() => databaseConfigs[props.databaseKey]?.icon || 'mdi:database')
</script>
