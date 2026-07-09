<template>
  <div class="bg-base-200 rounded-lg p-4 border border-base-300">
    <!-- Header -->
    <div class="flex items-center mb-3">
      <Icon :name="runtime.name === 'php' ? 'vscode-icons:file-type-php' : runtime.name === 'node' ? 'vscode-icons:file-type-node' : runtime.name === 'python' ? 'vscode-icons:file-type-python' : runtime.name === 'golang' ? 'vscode-icons:file-type-go' : runtime.name === 'redis' ? 'logos:redis' : runtime.name === 'mariadb' ? 'logos:mariadb' : runtime.name === 'mysql' ? 'logos:mysql' : 'mdi:package-variant'" 
            class="w-6 h-6 mr-3" 
            :class="runtime.available ? 'text-success' : 'text-base-content/40'" />
      <div class="flex-1">
        <h3 class="font-medium text-base-content capitalize">{{ runtime.name === 'php' ? 'PHP' : runtime.name === 'node' ? 'Node.js' : runtime.name === 'python' ? 'Python' : runtime.name === 'golang' ? 'Go' : runtime.name === 'redis' ? 'Redis' : runtime.name === 'mariadb' ? 'MariaDB' : runtime.name === 'mysql' ? 'MySQL' : runtime.name }}</h3>
        <p class="text-sm" :class="runtime.available ? 'text-success' : 'text-error'">{{ runtime.available ? 'Available' : 'Not installed / Inactive' }}</p>
      </div>
      <div class="flex items-center gap-2">
        <div :class="runtime.available ? 'badge badge-success badge-sm' : 'badge badge-error badge-sm'">
          {{ runtime.available ? 'Available' : 'Required' }}
        </div>
      </div>
    </div>

    <!-- Version Info -->
    <div class="text-xs text-base-content/60 mb-3">
      <div v-if="runtime.version">v{{ runtime.version }}</div>
      <div v-else>Not installed</div>
      <div v-if="!runtime.available && runtime.required">Minimum: v{{ runtime.required }}</div>
    </div>

    <!-- Install Button -->
    <button v-if="!runtime.available" 
            @click="installRuntime" 
            :disabled="installing"
            class="btn btn-sm btn-primary">
      <Icon v-if="installing" name="mdi:loading" class="animate-spin" />
      <Icon v-else name="mdi:download" />
      Install
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  runtime: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['installed'])

const installing = ref(false)

const installRuntime = async () => {
  installing.value = true
  try {
    const version = props.runtime.recommended || props.runtime.required
    const isDb = props.runtime.type === 'database' || ['mysql', 'mariadb', 'postgres', 'redis', 'mongodb', 'sqlite'].includes(props.runtime.name.toLowerCase())
    const baseEndpoint = isDb ? 'databases' : 'languages'
    const url = version === 'latest' 
      ? `/${baseEndpoint}/${props.runtime.name}/install`
      : `/${baseEndpoint}/${props.runtime.name}/install?version=${encodeURIComponent(version)}`
    
    const response = await useNuxtApp().$axiosApi.post(url)
    
    if (!response.data.success) {
      throw new Error(response.data.message || `Failed to install ${props.runtime.name}`)
    }
    
    useToaster(`${props.runtime.name} installed successfully`, "bg-green-500 text-white")
    emit('installed', { 
      type: isDb ? 'database' : 'runtime', 
      name: props.runtime.name,
      version: version
    })
  } catch (error) {
    const errorMessage = error.response?.data?.message || error.message || `Failed to install ${props.runtime.name}`
    useToaster(errorMessage, "bg-red-500 text-white")
  } finally {
    installing.value = false
  }
}
</script>
