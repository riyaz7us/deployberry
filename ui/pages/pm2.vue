<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold text-base-content">Process Manager (PM2)</h1>

      <!-- PM2 Installation Status -->
      <div v-if="!pm2Installed && !loadingInstallCheck" class="card p-4">
        <div class="flex items-center gap-2 text-error/70">
          <Icon name="mdi:alert-circle" size="20" />
          <span>PM2 is not installed on this system</span>
        </div>
        <p class="text-sm text-error/60 mt-2">
          Please install PM2 globally: <code class="bg-error/20 px-2 py-1 rounded">npm install -g pm2</code>
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="loadingInstallCheck" class="flex items-center gap-2 text-base-content/70">
        <Icon name="mdi:loading" size="20" class="animate-spin" />
        <span>Checking PM2 installation...</span>
      </div>

      <!-- PM2 Controls -->
      <div v-if="pm2Installed" class="space-y-4">
        <!-- Global Controls -->
        <div class="flex flex-wrap gap-3">
          <button @click="refreshProcesses" class="px-4 py-2 rounded btn btn-primary flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:refresh" size="16" />
            Refresh
          </button>
          <button @click="resurrectSaved" class="px-4 py-2 rounded btn btn-primary btn-outline flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:refresh" size="16" />
            Resurrect
          </button>
          <button @click="saveProcesses" class="px-4 py-2 rounded btn btn-primary btn-outline flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:floppy" size="16" />
            Save
          </button>
          <button @click="startAllProcesses" class="px-4 py-2 rounded btn btn-success flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:play" size="16" />
            Start All
          </button>
          <button @click="stopAllProcesses" class="px-4 py-2 rounded btn btn-error flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:stop" size="16" />
            Stop All
          </button>
          <button @click="restartAllProcesses" class="px-4 py-2 rounded btn btn-warning flex items-center gap-2"
            :disabled="loading">
            <Icon name="mdi:restart" size="16" />
            Restart All
          </button>
          <button @click="showCreateDialog = true" class="px-4 py-2 rounded btn btn-secondary flex items-center gap-2">
            <Icon name="mdi:plus" size="16" />
            Create Process
          </button>
        </div>

        <!-- Process List -->
        <div class="space-y-3">
          <h2 class="text-lg font-medium text-base-content">Running Processes</h2>

          <div v-if="loading && processes.length === 0" class="flex items-center justify-center py-8">
            <Icon name="mdi:loading" size="24" class="animate-spin text-base-content" />
            <span class="ml-2 text-base-content">Loading processes...</span>
          </div>

          <div v-else-if="processes.length === 0" class="text-center py-8 text-base-content/60">
            No PM2 processes found
          </div>

          <div v-else class="space-y-2">
            <div v-for="process in processes" :key="process.pm_id"
              class="p-4 rounded border border-base-300 bg-base-200">
              <div
                class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-3 pb-3 border-b border-base-300/30 sm:border-b-0 sm:pb-0">
                <div class="flex items-center gap-3 min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <div :class="[
                      'w-3 h-3 rounded-full flex-shrink-0',
                      getStatusColor(process.pm2_env?.status)
                    ]"></div>
                    <span class="text-base-content font-medium truncate">{{ process.name }}</span>
                    <span class="text-xs px-2 py-0.5 rounded bg-base-300 font-mono">
                      ID: {{ process.pm_id }}
                    </span>
                  </div>
                </div>

                <div class="flex gap-2 justify-end w-full sm:w-auto flex-wrap">
                  <button v-if="process.pm2_env?.status !== 'online'" @click="startProcess(process)"
                    class="text-xs px-3 py-1 rounded btn btn-success" :disabled="loading">
                    Start
                  </button>
                  <button v-if="process.pm2_env?.status === 'online'" @click="stopProcess(process)"
                    class="text-xs px-3 py-1 rounded btn btn-error" :disabled="loading">
                    Stop
                  </button>
                  <button @click="restartProcess(process)" class="text-xs px-3 py-1 rounded btn btn-warning"
                    :disabled="loading">
                    Restart
                  </button>
                  <button @click="deleteProcess(process)" class="text-xs px-3 py-1 rounded btn btn-neutral"
                    :disabled="loading">
                    Delete
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
                <div>
                  <span class="text-base-content/60">Status:</span>
                  <span :class="['ml-2 font-medium', getStatusTextColor(process.pm2_env?.status)]">
                    {{ process.pm2_env?.status || 'unknown' }}
                  </span>
                </div>
                <div>
                  <span class="text-base-content/60">CPU:</span>
                  <span class="ml-2 text-base-content">{{ process.monit?.cpu || 0 }}%</span>
                </div>
                <div>
                  <span class="text-base-content/60">Memory:</span>
                  <span class="ml-2 text-base-content">{{ formatMemory(process.monit?.memory) }}</span>
                </div>
                <div>
                  <span class="text-base-content/60">Uptime:</span>
                  <span class="ml-2 text-base-content">{{ formatUptime(process.pm2_env?.pm_uptime) }}</span>
                </div>
              </div>

              <div v-if="process.pm2_env?.exec_cwd" class="mt-2 text-sm">
                <span class="text-base-content/60">Path:</span>
                <span class="ml-2 text-base-content/70 font-mono">{{ process.pm2_env.exec_cwd }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Process Dialog -->
    <NativeDialog v-model="showCreateDialog">
      <div class="space-y-4 w-full max-w-md">
        <h2 class="text-base-content font-semibold">Create New PM2 Process</h2>

        <div class="space-y-3">
          <NativeTextField v-model="newProcess.name" label="Process Name" placeholder="my-app" />

          <NativeTextField v-model="newProcess.script" label="Script Path" placeholder="/path/to/app.js" />

          <NativeTextField v-model="newProcessArgs" label="Arguments (optional)"
            placeholder="--port 3000 --env production" hint="Separate arguments with spaces" />
        </div>

        <div class="flex gap-3 pt-4">
          <button @click="createProcess" class="px-4 py-2 rounded btn btn-success"
            :disabled="!newProcess.name || !newProcess.script || loading">
            Create Process
          </button>
          <button @click="showCreateDialog = false" class="px-4 py-2 rounded btn btn-neutral">
            Cancel
          </button>
        </div>
      </div>
    </NativeDialog>
  </div>
</template>

<script setup>
useHead({ title: 'Process Manager (PM2)' })
const pm2Installed = ref(false);
const loadingInstallCheck = ref(true);
const loading = ref(false);
const processes = ref([]);
const showCreateDialog = ref(false);

const newProcess = ref({
  name: '',
  script: '',
  args: []
});
const newProcessArgs = ref('');

onMounted(() => {
  checkPm2Installation();
});

async function checkPm2Installation() {
  loadingInstallCheck.value = true;
  try {
    const response = await useNuxtApp().$axiosApi.get('/pm2/installed');
    pm2Installed.value = response.data.success;

    if (pm2Installed.value) {
      await loadProcesses();
    }
  } catch (error) {
    console.error('Error checking PM2 installation:', error);
    useToaster('Failed to check PM2 installation', 'bg-error text-error-content');
  } finally {
    loadingInstallCheck.value = false;
  }
}

async function loadProcesses() {
  loading.value = true;
  try {
    const response = await useNuxtApp().$axiosApi.get('/pm2/list');
    processes.value = Array.isArray(response.data) ? response.data : [];
  } catch (error) {
    console.error('Error loading processes:', error);
    useToaster('Failed to load PM2 processes', 'bg-error text-error-content');
    processes.value = [];
  } finally {
    loading.value = false;
  }
}

async function refreshProcesses() {
  await loadProcesses();
  useToaster('Process list refreshed', 'bg-success text-success-content');
}

async function saveProcesses() {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/save');
    useToaster('All processes saved', 'bg-success text-success-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error saving processes:', error);
    useToaster('Failed to save processes', 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function resurrectSaved() {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/resurrect');
    useToaster('All processes resurrected', 'bg-success text-success-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error resurrecting processes:', error);
    useToaster('Failed to resurrect processes', 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function startAllProcesses() {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/start');
    useToaster('All processes started', 'bg-success text-success-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error starting all processes:', error);
    useToaster('Failed to start all processes', 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function stopAllProcesses() {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/stop');
    useToaster('All processes stopped', 'bg-warning text-warning-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error stopping all processes:', error);
    useToaster('Failed to stop all processes', 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function restartAllProcesses() {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/restart');
    useToaster('All processes restarted', 'bg-info text-info-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error restarting all processes:', error);
    useToaster('Failed to restart all processes', 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function startProcess(process) {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/process/start', {
      name: process.name,
      script: process.pm2_env?.pm_exec_path || process.name
    });
    useToaster(`Process ${process.name} started`, 'bg-success text-success-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error starting process:', error);
    useToaster(`Failed to start process ${process.name}`, 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function stopProcess(process) {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/process/stop', {
      name: process.name,
      script: process.pm2_env?.pm_exec_path || process.name
    });
    useToaster(`Process ${process.name} stopped`, 'bg-warning text-warning-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error stopping process:', error);
    useToaster(`Failed to stop process ${process.name}`, 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function restartProcess(process) {
  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/process/restart', {
      name: process.name,
      script: process.pm2_env?.pm_exec_path || process.name
    });
    useToaster(`Process ${process.name} restarted`, 'bg-info text-info-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error restarting process:', error);
    useToaster(`Failed to restart process ${process.name}`, 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function deleteProcess(process) {
  if (!confirm(`Are you sure you want to delete process "${process.name}"?`)) {
    return;
  }

  loading.value = true;
  try {
    await useNuxtApp().$axiosApi.post('/pm2/process/delete', {
      name: process.name,
      script: process.pm2_env?.pm_exec_path || process.name
    });
    useToaster(`Process ${process.name} deleted`, 'bg-neutral text-neutral-content');
    await loadProcesses();
  } catch (error) {
    console.error('Error deleting process:', error);
    useToaster(`Failed to delete process ${process.name}`, 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

async function createProcess() {
  loading.value = true;
  try {
    const processConfig = {
      name: newProcess.value.name,
      script: newProcess.value.script,
      args: newProcessArgs.value ? newProcessArgs.value.split(' ').filter(arg => arg.trim()) : []
    };

    await useNuxtApp().$axiosApi.post('/pm2/process/create', processConfig);
    useToaster(`Process ${newProcess.value.name} created`, 'bg-success text-success-content');

    // Reset form
    newProcess.value = { name: '', script: '', args: [] };
    newProcessArgs.value = '';
    showCreateDialog.value = false;

    await loadProcesses();
  } catch (error) {
    console.error('Error creating process:', error);
    useToaster(`Failed to create process ${newProcess.value.name}`, 'bg-error text-error-content');
  } finally {
    loading.value = false;
  }
}

function getStatusColor(status) {
  switch (status) {
    case 'online': return 'bg-success';
    case 'stopped': return 'bg-error';
    case 'stopping': return 'bg-warning';
    case 'launching': return 'bg-info';
    case 'errored': return 'bg-error';
    default: return 'bg-neutral';
  }
}

function getStatusTextColor(status) {
  switch (status) {
    case 'online': return 'text-success';
    case 'stopped': return 'text-error';
    case 'stopping': return 'text-warning';
    case 'launching': return 'text-info';
    case 'errored': return 'text-error';
    default: return 'text-base-content';
  }
}

function formatMemory(bytes) {
  if (!bytes) return '0 MB';
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
}

function formatUptime(timestamp) {
  if (!timestamp) return 'N/A';
  const now = Date.now();
  const uptime = now - timestamp;
  const seconds = Math.floor(uptime / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}
</script>

<style scoped>
/* Custom styles for PM2 interface */
.toast {
  position: fixed;
  top: 20px;
  right: 20px;
  padding: 12px 16px;
  border-radius: 6px;
  z-index: 1000;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
</style>
