<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold text-base-content flex items-center gap-3">
          <Icon name="mdi:web-plus" size="32" class="text-info/80" />
        Caddy Manager
      </h1>
      <p class="text-base-content/60">Manage caddy web server configurations and services</p>
    </div>
    <!-- Server Status Section -->
    <div class="card p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base-content font-semibold flex items-center gap-2">
          <Icon name="mdi:server" size="20" />
          Server Status
        </h2>
        <div class="flex items-center gap-2">
          <div :class="['w-3 h-3 rounded-full', statusIndicatorClass]"></div>
          <span class="text-sm text-base-content/70">
            {{ statusLabel }}
          </span>
          <span v-if="version" class="text-xs text-base-content/40">v{{ version }}</span>
        </div>
      </div>

      <div v-if="isInstalled" class="flex gap-3 flex-wrap">
        <!-- Start button shown only when inactive -->
        <button v-if="!isActive" @click="controlCaddy('start')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-success rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:play" size="16" />
          <span v-if="isLoading && actionType === 'start'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Start
        </button>
        <!-- Stop button shown only when active -->
        <button v-if="isActive" @click="controlCaddy('stop')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-error rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:stop" size="16" />
          <span v-if="isLoading && actionType === 'stop'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Stop
        </button>
        <!-- Reload and Restart always shown when installed -->
        <button @click="controlCaddy('reload')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-primary rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:refresh" size="16" />
          <span v-if="isLoading && actionType === 'reload'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Reload
        </button>
        <button @click="controlCaddy('restart')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-warning rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:restart" size="16" />
          <span v-if="isLoading && actionType === 'restart'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Restart
        </button>
      </div>

      <div v-else class="p-4 bg-warning/20 border border-warning/50 rounded-md">
        <div class="flex items-center gap-2 text-warning mb-2">
          <Icon name="mdi:information-outline" size="16" />
          <span class="text-sm font-medium">Caddy Not Installed</span>
        </div>
        <p class="text-xs text-warning/60">
          Caddy web server is not installed on this system. Please install it first to manage configurations.
        </p>
      </div>
    </div>

          <!-- Configuration Management Section -->
    <div class="rounded-lg p-6 bg-base-200">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base-content font-semibold flex items-center gap-2">
          <Icon name="mdi:file-document-multiple" size="20" />
          Configuration Files
        </h2>
        <button @click="showCreateConfig = true"
                class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors flex items-center gap-2">
          <Icon name="mdi:plus" size="16" />
          New Config
        </button>
      </div>

      <div v-if="configs.length === 0" class="text-center py-8">
        <Icon name="mdi:file-document-outline" size="48" class=" mx-auto mb-4" />
        <p class="text-base-content/60">No configuration files found</p>
        <p class="text-sm text-base-content/40 mt-2">Create your first caddy configuration file</p>
      </div>

      <div v-else class="space-y-3">
        <div v-for="config in configs"
             :key="config.domain"
             class="p-4 border border-base-content rounded-lg hover:bg-base-300 transition-colors">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="flex items-center gap-3 min-w-0">
              <Icon name="mdi:file-document" class="text-base-content/60 flex-shrink-0" size="20" />
              <div class="min-w-0">
                <h3 class="text-base-content font-medium truncate">{{ config.domain }}</h3>
                <p class="text-xs text-base-content/60">Created {{ formatDate(config.created_at) }}</p>
              </div>
            </div>
            <div class="flex items-center justify-between sm:justify-end gap-3 w-full sm:w-auto border-t border-base-content/20 pt-3 sm:pt-0 sm:border-t-0">
              <span class="px-2 py-1 rounded-full text-xs font-medium bg-success text-success-content">
                Active
              </span>
              <div class="flex gap-2">
                <button @click="openEditor(config)"
                        class="p-2 text-base-content/60 hover:text-blue-400 hover:bg-base-300 rounded transition-colors"
                        title="Edit Configuration">
                  <Icon name="mdi:pencil" size="16" />
                </button>
                <button @click="deleteConfig(config)"
                        class="p-2 text-base-content/60 hover:text-red-400 hover:bg-base-300 rounded transition-colors"
                        title="Delete Configuration">
                  <Icon name="mdi:delete" size="16" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

        <!-- Create Config Modal -->
    <NativeDialog v-model="showCreateConfig">
      <div v-if="showCreateConfig" class="space-y-6">
        <div class="flex items-center justify-between">
          <h3 class="text-white font-semibold flex items-center gap-2">
            <Icon name="mdi:plus" size="20" />
            Create New Caddy Configuration
          </h3>
        </div>

        <div class="space-y-4">
          <NativeTextField
            v-model="newConfigDomain"
            label="Domain"
            placeholder="example.com"
            :monospace="true"
          />

          <NativeTextField
            v-model="newConfigContent"
            label="Configuration Content"
            :textarea="true"
            :rows="12"
            :monospace="true"
            placeholder="example.com {
    root * /var/www/html
    file_server
}"
          />
        </div>

        <div class="flex justify-end gap-3">
          <button @click="showCreateConfig = false"
                  class="px-4 py-2 text-base-content/60 hover:text-white transition-colors">
            Cancel
          </button>
          <button @click="createConfig"
                  class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors">
            Create Configuration
          </button>
        </div>
      </div>
    </NativeDialog>

    <!-- Editor Modal -->
    <NativeDialog v-model="showEditor">
      <div v-if="showEditor" class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-white font-semibold flex items-center gap-2">
            <Icon name="mdi:pencil" size="20" />
            Edit Caddy Configuration
          </h3>
        </div>

        <div class="h-[60vh]">
          <CodeEditor
            v-if="editingConfig"
            :file-path="`/etc/caddy/sites-available/${editingConfig.domain}`"
            :is-open="showEditor"
            @close="closeEditor"
            @save="loadConfigs"
          />
        </div>
      </div>
    </NativeDialog>
  </div>
</template>

<script setup>
useHead({
  title: "Caddy Manager",
  link: [
    { rel: "stylesheet", href: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/default.min.css" }
  ],
  script: [
    { src: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js" }
  ]
})

import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'

const snackbar = useSnackbar()

// State variables
const configs = ref([])
const caddyStatus = ref(null)
const isInstalled = ref(false)
const isActive = ref(false)
const version = ref('')
const showCreateConfig = ref(false)
const newConfigDomain = ref('')
const newConfigContent = ref('')
const showEditor = ref(false)
const editingConfig = ref(null)
const isLoading = ref(false)
const actionType = ref(null)

// Computed status label and indicator class
const statusLabel = computed(() => {
  if (!isInstalled.value) return 'Caddy Not Installed'
  return isActive.value ? 'Caddy Running' : 'Caddy Stopped'
})

const statusIndicatorClass = computed(() => {
  if (!isInstalled.value) return 'bg-error'
  return isActive.value ? 'bg-success' : 'bg-gray-400'
})

const loadConfigs = () => {
  useNuxtApp()
    .$axiosApi.get('/webconfigs')
    .then((res) => {
      if (res.data.success) {
        // Filter only caddy configs or create from webconfigs
        configs.value = res.data.configs || []
      }
    })
    .catch((error) => {
      console.error('Error loading configs:', error)
    })
}

const checkCaddyStatus = () => {
  useNuxtApp()
    .$axiosApi.get('/caddy/status')
    .then((res) => {
      if (res.data.success) {
        caddyStatus.value = res.data
        isInstalled.value = res.data.installed || false
        isActive.value = res.data.active || false
        version.value = res.data.version || ''
      }
    })
    .catch((error) => {
      console.error('Error checking caddy status:', error)
    })
}

// Single unified status check replaces separate checkCaddyInstalled
const checkCaddyInstalled = checkCaddyStatus

// Caddy control functions
const controlCaddy = (action) => {
  isLoading.value = true
  actionType.value = action

  useNuxtApp()
    .$axiosApi.post(`/caddy/${action}`)
    .then(() => {
      checkCaddyStatus()
      snackbar.add({
        type: "success",
        text: `Caddy ${action} action completed successfully`
      })
    })
    .catch((error) => {
      console.error(`Error during caddy ${action}:`, error)
      snackbar.add({
        type: "error",
        text: `Error during caddy ${action}: ${error.response?.data?.message || error.response?.data?.error || error.message}`
      })
    })
    .finally(() => {
      isLoading.value = false
      actionType.value = null
    })
}

// Config management functions
const createConfig = () => {
  if (!newConfigDomain.value || !newConfigContent.value) return

  // This would need a backend endpoint for creating caddy configs
  // For now, we'll create a basic structure
  const configData = {
    domain: newConfigDomain.value,
    root_path: '/var/www/html',
    php_version: '',
    enable_gzip: true,
    enable_cache: false
  }

  useNuxtApp()
    .$axiosApi.post('/webconfigs', configData)
    .then(() => {
      showCreateConfig.value = false
      newConfigDomain.value = ''
      newConfigContent.value = ''
      loadConfigs()
    })
    .catch((error) => {
      console.error('Error creating config:', error)
    })
}

const deleteConfig = (config) => {
  if (!confirm(`Are you sure you want to delete the configuration for ${config.domain}?`)) return

  useNuxtApp()
    .$axiosApi.delete(`/webconfigs/${config.domain}`)
    .then(() => {
      loadConfigs()
    })
    .catch((error) => {
      console.error('Error deleting config:', error)
    })
}

const openEditor = (config) => {
  editingConfig.value = config
  showEditor.value = true
}

const closeEditor = () => {
  showEditor.value = false
  editingConfig.value = null
}

const formatDate = (dateString) => {
  if (!dateString) return 'Unknown'
  try {
    return dayjs(dateString).format('YYYY-MM-DD HH:mm:ss')
  } catch {
    return dateString
  }
}

// Load data on component mount
onMounted(() => {
  checkCaddyStatus()
  loadConfigs()
})
</script>