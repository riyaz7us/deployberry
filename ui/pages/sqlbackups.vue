<template>
  <div class="min-h-screen bg-base-100">
    <main class="container mx-auto px-4 py-6">
      <div class="bg-base-100 border border-base-300 rounded-lg p-6 mb-6">
        <div class="border-b border-base-300 pb-4 mb-4">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
            <h2 class="text-xl font-bold text-base-content">Backups for {{ database }}</h2>
            <div class="flex flex-wrap gap-2 w-full sm:w-auto">
              <button @click="createBackup"
                      class="btn btn-primary px-4 py-2 rounded inline-flex items-center text-sm">
                <Icon name="mdi:backup" class="mr-2" />
                <span>Create Backup</span>
              </button>
              <button @click="showCronModal = true"
                      class="btn btn-success px-4 py-2 rounded inline-flex items-center text-sm">
                <Icon name="mdi:clock-outline" class="mr-2" />
                <span>Schedule Backup</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Backup List -->
        <div class="hidden sm:grid grid-cols-12 gap-4 p-4 font-semibold border-b border-base-300 bg-base-200">
          <div class="text-base-content col-span-5">Name</div>
          <div class="text-base-content col-span-2">Created</div>
          <div class="text-base-content col-span-2">Modified</div>
          <div class="text-base-content col-span-1">Size</div>
          <div class="text-base-content col-span-2 text-right">Actions</div>
        </div>

        <div v-for="backup in backups"
             :key="backup.path"
             class="grid grid-cols-12 gap-2 sm:gap-4 p-3 sm:p-4 border-b border-base-300 hover:bg-base-300 items-center">
          <div class="col-span-8 sm:col-span-5 flex items-center min-w-0">
            <Icon name="mdi:file-document" class="mr-3 text-primary flex-shrink-0" />
            <span class="truncate font-medium text-base-content">{{ backup.name }}</span>
          </div>
          <div class="hidden sm:block col-span-2 text-sm text-base-content/70">{{ formatDate(backup.created) }}</div>
          <div class="hidden sm:block col-span-2 text-sm text-base-content/70">{{ formatDate(backup.modified) }}</div>
          <div class="hidden sm:block col-span-1 text-sm text-base-content/70">{{ formatSize(backup.size) }}</div>
          <div class="col-span-4 sm:col-span-2 flex justify-end gap-2">
            <button @click="restoreBackup(backup)"
                    class="text-success hover:text-success/80 py-1 px-2 rounded hover:bg-base-200 transition-colors flex items-center gap-1 text-xs font-semibold"
                    title="Restore Backup">
              <Icon name="mdi:restore" />
              <span class="hidden md:inline">Restore</span>
            </button>
            <button @click="deleteBackup(backup)"
                    class="text-error hover:text-error/80 p-2 rounded hover:bg-base-200 transition-colors"
                    title="Delete Backup">
              <Icon name="mdi:delete" size="18" />
            </button>
          </div>
          <!-- Sub-info for mobile (shows size/dates below name on small screens) -->
          <div class="col-span-8 sm:hidden pl-8 text-xs text-base-content/50 flex flex-wrap gap-2">
            <span>Created: {{ formatDate(backup.created) }}</span>
            <span>•</span>
            <span>Size: {{ formatSize(backup.size) }}</span>
          </div>
        </div>
      </div>

      <!-- Scheduled Backups -->
      <div class="bg-base-100 border border-base-300 rounded-lg p-6">
        <div class="border-b border-base-300 pb-4 mb-4">
          <h2 class="text-xl font-bold text-base-content">Scheduled Backups</h2>
        </div>

        <div class="grid grid-cols-2 gap-4 p-4 font-semibold border-b border-base-300 bg-base-200">
          <div class="text-base-content">Cron Schedule</div>
          <div class="text-base-content">Actions</div>
        </div>

        <div v-for="cron in crons"
             :key="cron"
             class="grid grid-cols-2 gap-4 p-4 border-b border-base-300 hover:bg-base-300">
          <div class="flex items-center">
            <Icon name="mdi:clock" class="mr-3 text-base-content/60" />
            <span>{{ cron }}</span>
          </div>
          <div class="flex space-x-2">
            <button @click="deleteCron(cron)"
                    class="text-error hover:text-error/80 px-2 py-1 rounded">
              <Icon name="mdi:delete" />
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Schedule Backup Modal -->
    <div v-if="showCronModal"
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-base-100 border border-base-300 p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold text-base-content mb-4">Schedule Backup</h3>
        <NativeTextField
          v-model="cronSchedule"
          placeholder="Cron schedule (e.g., 0 0 * * *)"
          hint="Format: minute hour day month weekday"
          class="mb-4"
        />
        <div class="flex justify-end space-x-2">
          <button @click="showCronModal = false"
                  class="px-4 py-2 text-base-content/70 hover:text-base-content">
            Cancel
          </button>
          <button @click="scheduleBackup"
                  class="px-4 py-2 btn btn-primary">
            Schedule
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
useHead({ title: 'Database Backups' })
import { ref, onMounted } from 'vue'
const { $axiosApi } = useNuxtApp()
const route = useRoute()

const database = ref(route.query.database || '')
const backups = ref([])
const crons = ref([])
const showCronModal = ref(false)
const cronSchedule = ref('')

const loadBackups = async () => {
  try {
    const response = await $axiosApi.get(`/sql/backups?database=${database.value}`)
    if (response.data.success) {
      backups.value = response.data.backups
    }
  } catch (error) {
    console.error('Error loading backups:', error)
  }
}

const loadCrons = async () => {
  try {
    const response = await $axiosApi.get(`/sql/backup/crons?database=${database.value}`)
    if (response.data.success) {
      crons.value = response.data.crons
    }
  } catch (error) {
    console.error('Error loading crons:', error)
  }
}

const createBackup = async () => {
  try {
    await $axiosApi.post('/sql/backup/create', { database: database.value })
    await loadBackups()
  } catch (error) {
    console.error('Error creating backup:', error)
  }
}

const restoreBackup = async (backup) => {
  if (!confirm(`Are you sure you want to restore from backup ${backup.name}? This will overwrite the current database.`)) return

  try {
    await $axiosApi.post('/sql/backup/restore', {
      database: database.value,
      backup_file: backup.path
    })
  } catch (error) {
    console.error('Error restoring backup:', error)
  }
}

const deleteBackup = async (backup) => {
  if (!confirm(`Are you sure you want to delete backup ${backup.name}?`)) return

  try {
    await $axiosApi.post('/sql/backup/delete', {
      database: database.value,
      backup_file: backup.path
    })
    await loadBackups()
  } catch (error) {
    console.error('Error deleting backup:', error)
  }
}

const scheduleBackup = async () => {
  if (!cronSchedule.value) return

  try {
    await $axiosApi.post('/sql/backup/cron', {
      database: database.value,
      cron_string: cronSchedule.value
    })
    showCronModal.value = false
    cronSchedule.value = ''
    await loadCrons()
  } catch (error) {
    console.error('Error scheduling backup:', error)
  }
}

const deleteCron = async (cron) => {
  if (!confirm('Are you sure you want to delete this backup schedule?')) return

  try {
    await $axiosApi.delete(`/sql/backup/cron?database=${database.value}&cron=${cron}`)
    await loadCrons()
  } catch (error) {
    console.error('Error deleting cron:', error)
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleString()
}

const formatSize = (size) => {
  if (!size) return 'N/A'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`
}

onMounted(() => {
  if (database.value) {
    loadBackups()
    loadCrons()
  }
})
</script> 