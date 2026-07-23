<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold text-base-content flex items-center gap-3">
          <Icon name="skill-icons:nginx" size="32" class="text-success/80" />
        Nginx Manager
      </h1>
      <p class="text-base-content/60">Manage nginx web server configurations and services</p>
    </div>

    <!-- Server Status Section -->
    <div class="card p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base-content font-semibold flex items-center gap-2">
          <Icon name="mdi:server" size="20" />
          Server Status
        </h2>
        <div class="flex items-center gap-2">
          <div :class="['w-3 h-3 rounded-full', isInstalled ? 'bg-success' : 'bg-error']"></div>
          <span class="text-sm text-base-content/70">
            {{ isInstalled ? 'Nginx Installed' : 'Nginx Not Installed' }}
          </span>
          <span v-if="version" class="text-xs text-base-content/40">v{{ version }}</span>
        </div>
      </div>

      <div v-if="isInstalled" class="flex gap-3 flex-wrap">
        <button @click="controlNginx('start')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-success rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:play" size="16" />
          <span v-if="isLoading && actionType === 'start'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Start
        </button>
        <button @click="controlNginx('stop')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-error rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:stop" size="16" />
          <span v-if="isLoading && actionType === 'stop'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Stop
        </button>
        <button @click="controlNginx('reload')"
                :disabled="isLoading"
                class="px-4 py-2 btn btn-primary rounded-md disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
          <Icon name="mdi:refresh" size="16" />
          <span v-if="isLoading && actionType === 'reload'">
            <Icon name="mdi:loading" size="16" class="animate-spin" />
          </span>
          Reload
        </button>
        <button @click="controlNginx('restart')"
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
          <span class="text-sm font-medium">Nginx Not Installed</span>
        </div>
        <p class="text-xs text-warning/60">
          Nginx web server is not installed on this system. Please install it first to manage configurations.
        </p>
      </div>
    </div>

    <!-- Configuration Management Section -->
    <div class="border border-base-300 rounded-lg p-6 bg-base-200">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-base-content font-semibold flex items-center gap-2">
          <Icon name="mdi:file-document-multiple" size="20" />
          Configuration Files
        </h2>
        <button @click="showCreateConfig = true"
                class="px-4 py-2 btn btn-primary rounded-md flex items-center gap-2">
          <Icon name="mdi:plus" size="16" />
          New Config
        </button>
      </div>

      <div v-if="configs.length === 0" class="text-center py-8">
        <Icon name="mdi:file-document-outline" size="48" class="text-slate-600 mx-auto mb-4" />
        <p class="text-base-content/60">No configuration files found</p>
        <p class="text-sm text-base-content/40 mt-2">Create your first nginx configuration file</p>
      </div>

      <div v-else class="space-y-3">
        <div v-for="config in configs"
             :key="config.name"
             class="p-4 border border-base-300 bg-base-200/50 rounded-lg hover:bg-base-200 transition-colors">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="flex items-center gap-3 min-w-0">
              <Icon name="mdi:file-document" class="text-base-content/60 flex-shrink-0" size="20" />
              <div class="min-w-0">
                <h3 class="text-base-content font-medium truncate">{{ config.name }}</h3>
                <p class="text-xs text-base-content/60">Created {{ formatDate(config.created_at) }}</p>
              </div>
            </div>
            <div class="flex items-center justify-between sm:justify-end gap-3 w-full sm:w-auto border-t border-base-300/30 pt-3 sm:pt-0 sm:border-t-0">
              <span :class="['px-2 py-1 rounded-full text-xs font-medium',
                            config.enabled ? 'bg-success text-success-content' : 'bg-error text-error-content']">
                {{ config.enabled ? 'Enabled' : 'Disabled' }}
              </span>
              <div class="flex gap-2">
                <button @click="openEditor(config)"
                        class="p-2 text-base-content/60 hover:text-primary/80 hover:bg-base-300 rounded transition-colors"
                        title="Edit Configuration">
                  <Icon name="mdi:pencil" size="16" />
                </button>
                <button @click="deleteConfig(config)"
                        class="p-2 text-base-content/60 hover:text-error/80 hover:bg-base-300 rounded transition-colors"
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
            Create New Nginx Configuration
          </h3>
        </div>

        <div class="space-y-4">
          <NativeTextField
            v-model="newConfigName"
            label="Configuration Name"
            placeholder="example.com"
            :monospace="true"
          />

          <NativeTextField
            v-model="newConfigContent"
            label="Configuration Content"
            :textarea="true"
            :rows="12"
            :monospace="true"
            placeholder="server {
    listen 80;
    server_name example.com;

    location / {
        root /var/www/html;
        index index.html;
    }
}"
          />
        </div>

        <div class="flex justify-end gap-3">
          <button @click="showCreateConfig = false"
                  class="px-4 py-2 text-base-content/60 hover:text-white transition-colors">
            Cancel
          </button>
          <button @click="createConfig"
                  class="px-4 py-2 btn btn-primary rounded-md">
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
            Edit Nginx Configuration
          </h3>
        </div>

        <div class="h-[60vh]">
          <CodeEditor
            v-if="editingConfig"
            :file-path="editingConfig.name"
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
  title: "Nginx Manager",
  link: [
    { rel: "stylesheet", href: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/default.min.css" }
  ],
  script: [
    { src: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js" },
    { src: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/languages/nginx.min.js" }
  ]
})

import { ref, onMounted } from 'vue'
import dayjs from 'dayjs'



// State variables
const configs = ref([])
const nginxStatus = ref(null)
const isInstalled = ref(false)
const version = ref('')
const showCreateConfig = ref(false)
const newConfigName = ref('')
const newConfigContent = ref('')
const showEditor = ref(false)
const editingConfig = ref(null)
const isLoading = ref(false)
const actionType = ref(null)

const loadConfigs = () => {
  useNuxtApp()
    .$axiosApi.get('/nginx/configs')
    .then((res) => {
      if (res.data.success) {
        configs.value = res.data.configs
      }
    })
    .catch((error) => {
      console.error('Error loading configs:', error)
    })
}

const checkNginxStatus = () => {
  useNuxtApp()
    .$axiosApi.get('/nginx/status')
    .then((res) => {
      nginxStatus.value = res.data
    })
    .catch((error) => {
      console.error('Error checking nginx status:', error)
    })
}

const checkNginxInstalled = () => {
  useNuxtApp()
    .$axiosApi.get('/nginx/installed')
    .then((res) => {
      isInstalled.value = res.data.success
    })
    .catch((error) => {
      console.error('Error checking nginx installation:', error)
    })
}

// Nginx control functions
const controlNginx = (action) => {
  isLoading.value = true
  actionType.value = action

  useNuxtApp()
    .$axiosApi.post(`/nginx/${action}`)
    .then(() => {
      checkNginxStatus()
    })
    .catch((error) => {
      console.error(`Error during nginx ${action}:`, error)
    })
    .finally(() => {
      isLoading.value = false
      actionType.value = null
    })
}

// Config management functions
const createConfig = () => {
  if (!newConfigName.value || !newConfigContent.value) return

  useNuxtApp()
    .$axiosApi.post('/nginx/config/add', {
      name: newConfigName.value,
      content: newConfigContent.value
    })
    .then(() => {
      showCreateConfig.value = false
      newConfigName.value = ''
      newConfigContent.value = ''
      loadConfigs()
    })
    .catch((error) => {
      console.error('Error creating config:', error)
    })
}

const deleteConfig = (config) => {
  if (!confirm(`Are you sure you want to delete ${config.name}?`)) return

  useNuxtApp()
    .$axiosApi.post('/nginx/config/delete', {
      name: config.name
    })
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

const formatDate = (timestamp) => {
  return dayjs(timestamp * 1000).format('YYYY-MM-DD HH:mm:ss')
}

// Load data on component mount
onMounted(() => {
  checkNginxInstalled()
  checkNginxStatus()
  loadConfigs()
})
</script>