<template>
  <div class="bg-base-200 rounded-lg p-4 border border-base-300">
    <!-- Header -->
    <div class="flex items-center mb-3">
      <Icon name="mdi:database" class="w-6 h-6 mr-3" 
            :class="database?.options && database.options.some(opt => opt.available) ? 'text-warning' : 'text-error'" />
      <div class="flex-1">
        <h3 class="font-medium text-base-content">Database</h3>
        <p class="text-sm" :class="database?.options && database.options.some(opt => opt.available) ? 'text-warning' : 'text-error'">
          {{ database?.options && database.options.some(opt => opt.available) ? 'Select database' : 'No compatible databases' }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <div :class="database?.options && database.options.some(opt => opt.available) ? 'badge badge-warning badge-sm' : 'badge badge-error badge-sm'">
          Required
        </div>
      </div>
    </div>

    <!-- Database Selection -->
    <div v-if="database && database.options && database.options.length > 0" class="space-y-2">
      <!-- Database Selection -->
      <select v-model="selectedDB" @change="emit('db-selected', selectedDB)" 
              class="select select-sm select-bordered w-full">
        <option value="">Select database...</option>
        <option v-for="option in database.options" :key="option.name" :value="option.name">
          {{ option.name === 'mysql' ? 'MySQL' : option.name === 'mariadb' ? 'MariaDB' : option.name === 'postgres' ? 'PostgreSQL' : option.name === 'mongodb' ? 'MongoDB' : option.name === 'redis' ? 'Redis' : option.name === 'sqlite' ? 'SQLite' : option.name }} 
          <span v-if="option.version">v{{ option.version }}</span>
          <span v-else>(Not installed)</span>
        </option>
      </select>

      <!-- Version Info for Selected DB -->
      <div v-if="selectedOption" class="text-xs text-base-content/60">
        <div v-if="selectedOption.version">v{{ selectedOption.version }}</div>
        <div v-else>Not installed</div>
        <div v-if="!selectedOption.available && selectedOption.required">Minimum: v{{ selectedOption.required }}</div>
      </div>

      <!-- Install Button for selected DB -->
      <button v-if="selectedDB && !selectedOption.available" 
              @click="installDatabase" 
              :disabled="installingDB"
              class="btn btn-sm btn-primary">
        <Icon v-if="installingDB" name="mdi:loading" class="animate-spin" />
        <Icon v-else name="mdi:download" />
        Install
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  database: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['installed', 'db-selected'])

const installingDB = ref(false)
const selectedDB = ref(props.database?.selected || '')

const selectedOption = computed(() => {
  if (!props.database?.options) return null
  return props.database.options.find(opt => opt.name === selectedDB.value)
})

const installDatabase = async () => {
  installingDB.value = true
  try {
    const version = selectedOption.value?.recommended || selectedOption.value?.required
    const url = version === 'latest' 
      ? `/databases/${selectedDB.value}/install`
      : `/databases/${selectedDB.value}/install?version=${encodeURIComponent(version)}`
    
    const response = await useNuxtApp().$axiosApi.post(url)
    
    if (!response.data.success) {
      throw new Error(response.data.message || `Failed to install ${selectedDB.value}`)
    }
    
    useToaster(`${selectedDB.value} installed successfully`, "bg-green-500 text-white")
    emit('installed', { 
      type: 'database', 
      name: selectedDB.value,
      version: version
    })
  } catch (error) {
    const errorMessage = error.response?.data?.message || error.message || `Failed to install ${selectedDB.value}`
    useToaster(errorMessage, "bg-red-500 text-white")
  } finally {
    installingDB.value = false
  }
}
</script>
