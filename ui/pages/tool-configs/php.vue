<template>
  <div class="px-6 py-4 space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold text-base-content">PHP Configuration</h1>
      <p class="text-base-content/60">Manage PHP settings, extensions, and php.ini configuration</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Icon name="mdi:loading" size="32" class="animate-spin text-base-content" />
      <span class="ml-3 text-base-content">Loading PHP configuration...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="border border-error bg-error/10 p-4 rounded text-error">
      {{ error }}
    </div>

    <!-- Configuration Content -->
    <div v-else-if="config" class="space-y-6">
      <!-- Quick Actions -->
      <div class="flex gap-3 flex-wrap">
        <button @click="refreshConfig"
          class="px-4 py-2 btn btn-primary rounded"
          :disabled="loading">
          <Icon name="mdi:refresh" size="16" class="inline mr-2" />
          Refresh
        </button>
        <button @click="openConfigDialog"
          class="px-4 py-2 btn btn-success rounded">
          <Icon name="mdi:cog" size="16" class="inline mr-2" />
          Edit Configuration
        </button>
        <button @click="openExtensionsDialog"
          class="px-4 py-2 btn btn-secondary rounded">
          <Icon name="mdi:puzzle" size="16" class="inline mr-2" />
          Manage Extensions
        </button>
      </div>

      <!-- Current Configuration Overview -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div class="bg-base-200 border border-base-300 rounded p-4">
          <h3 class="text-sm font-semibold text-base-content/70 mb-2">Memory & Execution</h3>
          <div class="space-y-1 text-sm">
            <div class="flex justify-between">
              <span class="text-base-content/60">Memory Limit:</span>
              <span class="text-base-content">{{ config.memory_limit || 'Not set' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Max Execution:</span>
              <span class="text-base-content">{{ config.max_execution_time || 'Not set' }}s</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Max Input Time:</span>
              <span class="text-base-content">{{ config.max_input_time || 'Not set' }}s</span>
            </div>
          </div>
        </div>

        <div class="bg-base-200 border border-base-300 rounded p-4">
          <h3 class="text-sm font-semibold text-base-content/70 mb-2">Upload Settings</h3>
          <div class="space-y-1 text-sm">
            <div class="flex justify-between">
              <span class="text-base-content/60">Post Max Size:</span>
              <span class="text-base-content">{{ config.post_max_size || 'Not set' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Upload Max:</span>
              <span class="text-base-content">{{ config.upload_max_filesize || 'Not set' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Max Files:</span>
              <span class="text-base-content">{{ config.max_file_uploads || 'Not set' }}</span>
            </div>
          </div>
        </div>

        <div class="bg-base-200 border border-base-300 rounded p-4">
          <h3 class="text-sm font-semibold text-base-content/70 mb-2">Error Handling</h3>
          <div class="space-y-1 text-sm">
            <div class="flex justify-between">
              <span class="text-base-content/60">Display Errors:</span>
              <span class="text-base-content">{{ config.display_errors || 'Not set' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Log Errors:</span>
              <span class="text-base-content">{{ config.log_errors || 'Not set' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Error Reporting:</span>
              <span class="text-base-content">{{ config.error_reporting || 'Not set' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Extensions Overview -->
      <div class="bg-base-200 border border-base-300 rounded p-4">
        <h3 class="text-sm font-semibold text-base-content/70 mb-3">Loaded Extensions</h3>
        <div class="flex flex-wrap gap-2">
          <span v-for="extension in enabledExtensions" :key="extension"
            class="text-xs px-2 py-1 rounded bg-success text-success-content">
            {{ extension }}
          </span>
        </div>
        <div class="mt-3 text-sm text-base-content/60">
          {{ enabledExtensionsCount }} extensions loaded
        </div>
      </div>

      <!-- PHP.ini Path -->
      <div v-if="iniPath" class="bg-base-200 border border-base-300 rounded p-4">
        <h3 class="text-sm font-semibold text-base-content/70 mb-2">Configuration File</h3>
        <div class="text-sm font-mono text-base-content/70 bg-base-200/50 p-2 rounded">
          {{ iniPath }}
        </div>
      </div>
    </div>

    <!-- Configuration Edit Dialog -->
    <NativeDialog v-model="configDialog">
      <div class="space-y-4">
        <h2 class="text-base-content font-semibold">Edit PHP Configuration</h2>
        
        <div v-if="savingConfig" class="flex items-center justify-center py-4">
          <Icon name="mdi:loading" size="24" class="animate-spin text-base-content" />
          <span class="ml-2 text-base-content">Saving configuration...</span>
        </div>
        
        <div v-else class="space-y-4 max-h-96 overflow-y-auto">
          <!-- Memory & Execution Settings -->
          <div class="space-y-3">
            <h3 class="text-sm font-semibold text-base-content/70">Memory & Execution</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <NativeTextField
                v-model="editConfig.memory_limit"
                label="Memory Limit"
                placeholder="128M"
              />
              <NativeTextField
                v-model="editConfig.max_execution_time"
                label="Max Execution Time (seconds)"
                placeholder="30"
              />
              <NativeTextField
                v-model="editConfig.max_input_time"
                label="Max Input Time (seconds)"
                placeholder="60"
              />
            </div>
          </div>

          <!-- Upload Settings -->
          <div class="space-y-3">
            <h3 class="text-sm font-semibold text-base-content/70">Upload Settings</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <NativeTextField
                v-model="editConfig.post_max_size"
                label="Post Max Size"
                placeholder="8M"
              />
              <NativeTextField
                v-model="editConfig.upload_max_filesize"
                label="Upload Max Filesize"
                placeholder="2M"
              />
              <NativeTextField
                v-model="editConfig.max_file_uploads"
                label="Max File Uploads"
                placeholder="20"
              />
            </div>
          </div>

          <!-- Error Handling -->
          <div class="space-y-3">
            <h3 class="text-sm font-semibold text-base-content/70">Error Handling</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-base-content/60 mb-1">Display Errors</label>
                <select v-model="editConfig.display_errors" 
                  class="w-full px-3 py-2 bg-base-200 border border-base-300 rounded text-base-content text-sm">
                  <option value="Off">Off</option>
                  <option value="On">On</option>
                </select>
              </div>
              <div>
                <label class="block text-xs text-base-content/60 mb-1">Log Errors</label>
                <select v-model="editConfig.log_errors" 
                  class="w-full px-3 py-2 bg-base-200 border border-base-300 rounded text-base-content text-sm">
                  <option value="Off">Off</option>
                  <option value="On">On</option>
                </select>
              </div>
              <div class="md:col-span-2">
                <label class="block text-xs text-base-content/60 mb-1">Error Reporting</label>
                <select v-model="editConfig.error_reporting" 
                  class="w-full px-3 py-2 bg-base-200 border border-base-300 rounded text-base-content text-sm">
                  <option value="E_ALL">E_ALL (All errors and warnings)</option>
                  <option value="E_ALL & ~E_DEPRECATED">E_ALL & ~E_DEPRECATED (All except deprecated)</option>
                  <option value="E_ALL & ~E_NOTICE">E_ALL & ~E_NOTICE (All except notices)</option>
                  <option value="E_ERROR | E_WARNING | E_PARSE">E_ERROR | E_WARNING | E_PARSE (Errors, warnings, and parse errors only)</option>
                  <option value="0">0 (No error reporting)</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Custom Settings -->
          <div class="space-y-3">
            <h3 class="text-sm font-semibold text-base-content/70">Custom Settings</h3>
            <div v-for="(value, key) in editConfig.custom_settings" :key="key" 
              class="flex gap-2 items-center">
              <NativeTextField
                :model-value="key"
                disabled
                class="flex-1"
              />
              <NativeTextField
                v-model="editConfig.custom_settings[key]"
                class="flex-1"
              />
              <button @click="deleteCustomSetting(key)"
                class="px-2 py-2 text-error hover:text-error/70">
                <Icon name="mdi:delete" size="16" />
              </button>
            </div>
            <button @click="addCustomSetting"
              class="text-xs px-3 py-2 btn btn-neutral rounded">
              Add Custom Setting
            </button>
          </div>
        </div>

        <div class="flex gap-3 justify-end pt-4 border-t border-slate-700">
          <button @click="configDialog = false" 
            class="px-4 py-2 text-base-content/70 hover:text-base-content transition-colors">
            Cancel
          </button>
          <button @click="saveConfig"
            class="px-4 py-2 btn btn-success rounded"
            :disabled="savingConfig">
            Save Changes
          </button>
        </div>
      </div>
    </NativeDialog>

    <!-- Extensions Management Dialog -->
    <NativeDialog v-model="extensionsDialog">
      <div class="space-y-4">
        <h2 class="text-base-content font-semibold">Manage PHP Extensions</h2>
        
        <div class="space-y-3 max-h-96 overflow-y-auto">
          <div v-for="(enabled, extension) in config?.extensions" :key="extension"
            class="flex items-center justify-between p-3 bg-base-200 border border-base-300 rounded">
            <div class="flex items-center gap-3">
              <span class="text-base-content">{{ extension }}</span>
              <span v-if="enabled" class="text-xs px-2 py-0.5 rounded bg-success text-success-content">Loaded</span>
              <span v-else class="text-xs px-2 py-0.5 rounded bg-base-300 text-base-content">Available</span>
            </div>
            <label class="flex items-center cursor-pointer">
              <input type="checkbox" :checked="enabled" @change="toggleExtension(extension, $event.target.checked)"
                class="sr-only">
              <div class="relative">
                <div :class="['block bg-base-300 w-10 h-6 rounded-full', enabled ? 'bg-success' : '']"></div>
                <div :class="['absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition transform', enabled ? 'translate-x-4' : '']"></div>
              </div>
            </label>
          </div>
        </div>

        <div class="text-xs text-base-content/60">
          Note: Extension changes require restarting your web server to take effect.
        </div>

        <div class="flex justify-end pt-4 border-t border-slate-700">
          <button @click="extensionsDialog = false" 
            class="px-4 py-2 text-base-content/70 hover:text-base-content transition-colors">
            Close
          </button>
        </div>
      </div>
    </NativeDialog>
  </div>
</template>

<script setup>
useHead({ title: 'PHP Configuration' })
const snackbar = useSnackbar()

// State
const loading = ref(false)
const error = ref(null)
const config = ref(null)
const iniPath = ref('')
const configDialog = ref(false)
const extensionsDialog = ref(false)
const savingConfig = ref(false)
const editConfig = ref({})

// Computed
const enabledExtensionsCount = computed(() => {
  if (!config.value?.extensions) return 0
  return Object.values(config.value.extensions).filter(Boolean).length
})

const enabledExtensions = computed(() => {
  if (!config.value?.extensions) return []
  return Object.keys(config.value.extensions).filter(ext => config.value.extensions[ext])
})

// Lifecycle
onMounted(() => {
  loadConfig()
})

// Methods
async function loadConfig() {
  loading.value = true
  error.value = null
  
  try {
    const response = await useNuxtApp().$axiosApi.get('/languages/php/config')
    config.value = response.data.data
    iniPath.value = response.data.ini_path
  } catch (err) {
    error.value = err.response?.data?.message || 'Failed to load PHP configuration'
    console.error('Failed to load PHP config:', err)
  } finally {
    loading.value = false
  }
}

function refreshConfig() {
  loadConfig()
}

function openConfigDialog() {
  editConfig.value = {
    memory_limit: config.value.memory_limit || '',
    max_execution_time: config.value.max_execution_time || '',
    max_input_time: config.value.max_input_time || '',
    post_max_size: config.value.post_max_size || '',
    upload_max_filesize: config.value.upload_max_filesize || '',
    max_file_uploads: config.value.max_file_uploads || '',
    display_errors: config.value.display_errors || 'Off',
    error_reporting: config.value.error_reporting || 'E_ALL',
    log_errors: config.value.log_errors || 'On',
    custom_settings: { ...config.value.custom_settings } || {}
  }
  configDialog.value = true
}

function openExtensionsDialog() {
  extensionsDialog.value = true
}

async function saveConfig() {
  savingConfig.value = true
  
  try {
    await useNuxtApp().$axiosApi.post('/languages/php/config', editConfig.value)
    
    snackbar.add({
      type: 'success',
      text: 'PHP configuration updated successfully'
    })
    
    configDialog.value = false
    await loadConfig() // Reload to get updated values
  } catch (err) {
    snackbar.add({
      type: 'error',
      text: err.response?.data?.message || 'Failed to update PHP configuration'
    })
    console.error('Failed to save PHP config:', err)
  } finally {
    savingConfig.value = false
  }
}

function toggleExtension(extension, enabled) {
  if (config.value?.extensions) {
    config.value.extensions[extension] = enabled
  }
  
  // Note: In a real implementation, you might want to call an API to actually enable/disable the extension
  snackbar.add({
    type: 'info',
    text: `Extension ${extension} ${enabled ? 'enabled' : 'disabled'}. Restart your web server for changes to take effect.`
  })
}

function addCustomSetting() {
  const key = prompt('Enter setting name:')
  if (key && key.trim()) {
    editConfig.value.custom_settings[key.trim()] = ''
  }
}

function deleteCustomSetting(key) {
  if (confirm(`Delete custom setting "${key}"?`)) {
    delete editConfig.value.custom_settings[key]
  }
}
</script>

<style scoped>
</style>