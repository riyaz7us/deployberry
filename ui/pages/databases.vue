<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold text-base-content">Database Servers</h1>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <DbPill v-for="dbKey in dbServers" :key="dbKey" :database-key="dbKey" @click="selectDB({key: dbKey})"
          :version="getVersionStatus(dbKey).version" 
          :installed="getVersionStatus(dbKey).isInstalled"
          :port="getVersionStatus(dbKey).port"
          :loadingOverview="loadingOverview" />
      </div>
    </div>

    <NativeDialog v-model="database">
      <NativeInstalling v-if="progressing" />
      <div v-else class="space-y-4">
        <h2 class="text-base-content font-semibold">Manage {{ database?.label }}</h2>
        <div v-if="database" class="space-y-3">
          <div class="flex items-center gap-3">
            <span class="text-slate-300">Available {{ database.label }} versions</span>
            <button @click="fetchVersions"
              class="text-xs px-2 py-1 border border-base-300 rounded hover:bg-base-300 transition-colors"
              :disabled="loadingVersions">
              Refresh
            </button>
            <button @click="runHardCheck"
              class="text-xs px-2 py-1 border border-primary text-primary rounded hover:bg-primary/30 transition-colors"
              :disabled="hardCheckLoading">
              <span v-if="!hardCheckLoading">Check System</span>
              <span v-else class="flex items-center">
                <Icon name="mdi:loading" size="14" class="animate-spin mr-1" /> Checking...
              </span>
            </button>
          </div>

          <div v-if="hardCheckResults" class="mt-4 p-3 bg-base-200 border border-base-300 rounded">
            <h3 class="text-sm font-bold text-slate-300 mb-2">System Detected Versions</h3>
            <p class="text-sm font-medium mb-2 italic text-red-500 max-w-3xl">Following are the system detected versions of your database. If these versions aren't registered below, or causing incompatibility issues, you need to uninstall them, If you already have the data loaded on this system, please backup your databases or contact your system administrator</p>
            <button @click="uninstallSystem()"
              class="text-xs px-2 py-1 rounded btn btn-warning"
              :disabled="loadingVersions">
              Uninstall
            </button>
            <div class="space-y-1">
              <div v-for="(version, key) in hardCheckResults" :key="key" class="flex justify-between text-sm">
                <span class="text-slate-400">{{ key }}:</span>
                <span class="font-mono text-slate-200">{{ version || 'Not found' }}</span>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <div v-show="warnings" class="border-2 border-warning bg-warning/10 p-2 text-warning">
              {{ warnings }}
            </div>
            <div v-if="loadingVersions" class="flex items-center justify-center py-4">
              <Icon name="mdi:loading" size="24" class="animate-spin text-base-content" />
              <span class="ml-2 text-base-content">Loading versions...</span>
            </div>

            <div v-else-if="versionEntries.length === 0" class="text-center py-4 text-slate-400">
              No versions available
            </div>
            <div v-for="e in versionEntries" :key="e.version" class="bg-base-200 border border-base-300 rounded p-3">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-2">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-slate-200 font-medium">{{ database.label }} {{ e.version }}</span>
                  <span v-if="e.active" class="text-xs px-2 py-0.5 rounded bg-success text-success-content font-medium">Active</span>
                  <span v-else-if="e.installed"
                    class="text-xs px-2 py-0.5 rounded bg-info text-info-content font-medium">Installed</span>
                </div>
                <div class="flex gap-2 justify-end w-full sm:w-auto border-t border-base-300/30 pt-2 sm:pt-0 sm:border-t-0">
                  <button v-if="e.installed && !e.active" @click="activateDB()"
                    class="text-xs px-2 py-1 rounded btn btn-primary btn-sm"
                    :disabled="loadingVersions">
                    Activate
                  </button>
                  <button v-else-if="e.installed && e.active" @click="deactivateDB()"
                    class="text-xs px-2 py-1 rounded btn btn-error btn-sm"
                    :disabled="loadingVersions">
                    Deactivate
                  </button>
                  <button v-if="e.installed" @click="uninstallDB()"
                    class="text-xs px-2 py-1 rounded btn btn-warning btn-sm"
                    :disabled="loadingVersions">
                    Uninstall
                  </button>
                  <button v-else-if="!e.installed" @click="installDB(e.version)"
                    class="text-xs px-2 py-1 rounded btn btn-neutral btn-sm"
                    :disabled="loadingVersions">
                    Install
                  </button>
                </div>
              </div>

              <div v-if="e.installed" class="mt-2 text-sm text-slate-300 space-y-1">
                <div class="flex items-center justify-between">
                  <span>Port:</span>
                  <span class="font-mono">{{ e.port || '3306' }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span>Root Password:</span>
                  <div class="flex items-center gap-1">
                    <span class="font-mono">{{ showPassword ? e.root_password : '••••••••' }}</span>
                    <button @click.stop="togglePassword" class="text-base-content/70 hover:text-base-content"
                      :title="showPassword ? 'Hide password' : 'Show password'">
                      <Icon :name="showPassword ? 'mdi:eye-off' : 'mdi:eye'" size="16" />
                    </button>
                    <button @click.stop="copyToClipboard(e.root_password)" class="text-base-content/70 hover:text-base-content"
                      title="Copy to clipboard">
                      <Icon name="mdi:content-copy" size="16" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </NativeDialog>
  </div>
</template>

<script setup>
const snackbar = useSnackbar()

const versionsState = ref({});
const showPassword = ref(false);

const dbServers = [
  "mysql", "mariadb", "postgres", "sqlite", "mongodb", "redis"
];

const dialog = ref(false);
const database = ref(null);
const versionEntries = ref([]);
const warnings = ref("");
const loadingOverview = ref(false);
const loadingVersions = ref(false);
const progressing = ref(false);
const hardCheckResults = ref(null)
const hardCheckLoading = ref(false)

onMounted(() => {
  loadOverview();
});

function loadOverview() {
  loadingOverview.value = true;
  useNuxtApp()
    .$axiosApi.get("/databases")
    .then((res) => {
      // Ensure we have an object even if the response is empty
      versionsState.value = res.data?.data || {};

      // Log for debugging
      console.log('Loaded database versions:', versionsState.value);
    })
    .catch((error) => {
      snackbar.add({
        type: "error",
        text: `Failed To Load DB Status`,
      });
    })
    .finally(() => {
      loadingOverview.value = false;
    });
}

function selectDB(dbData) {
  // Create full database object for dialog usage
  const dbKey = dbData.key;
  database.value = {
    key: dbKey,
    label: dbKey.charAt(0).toUpperCase() + dbKey.slice(1) // Simple fallback
  };
  versionEntries.value = []
  hardCheckResults.value = null
  dialog.value = true
  fetchVersions();
}

function fetchVersions() {
  if (!database.value) return;

  loadingVersions.value = true;
  useNuxtApp()
    .$axiosApi.get(`/databases/${database.value.key}/versions`)
    .then((res) => {
      versionEntries.value = Array.isArray(res.data.data) ? res.data.data : [];
      warnings.value = res.data.warnings;
    })
    .catch((error) => {
      snackbar.add({
        type: "error",
        text: `Failed To Fetch DB Versions`,
      });
      versionEntries.value = [];
    })
    .finally(() => {
      loadingVersions.value = false;
    });
}

async function installDB(version) {
  if (!database.value) return;
  console.log("versionEntries", versionEntries.value)
  if (versionEntries.value.find(e => e.installed)) {
    snackbar.add({
      type: "error",
      text: `Please uninstall other versions first`,
    });
    return;
  };
  progressing.value=true;
  useNuxtApp().$axiosApi.post(`/databases/${database.value.key}/install?version=${version}`)
    .then((res) => {
      loadOverview();
      fetchVersions();
      snackbar.add({
        type: "success",
        text: `${database.value.label} ${version} Installed successfully`,
      });
    },(err)=>{
      snackbar.add({
        type: "error",
        text: `Failed To Install ${database.value.label} ${version}`,
      });
      console.error('Installation error:', err);
    })
    .finally(() => {
      progressing.value = false;
    });
}

function activateDB() {
  progressing.value = true;
  if (!database.value) return;
  useNuxtApp().$axiosApi.post(`/databases/${database.value.key}/activate`).then((res) => {
    snackbar.add({
      type: "success",
      text: `${database.value.label} Activated successfully`,
    });
  }, (err) => {
    snackbar.add({
      type: "error",
      text: `Failed To Activate ${database.value.label}`,
    });
    console.error('Activation error:', err);
  }).finally(() => {
    // Refresh data
    fetchVersions();
    loadOverview();
    progressing.value = false;
  })
}

function deactivateDB() {
    if (!database.value) return;
  progressing.value = true;

  useNuxtApp().$axiosApi.post(`/databases/${database.value.key}/deactivate`).then((res) => {
    snackbar.add({
      type: "success",
      text: `${database.value.label} Deactivated Successfully`,
    });

  }, (err) => {
    snackbar.add({
      type: "error",
      text: `Failed To Deactivate ${database.value.label}`,
    });
    console.error('Deactivation error:', err);
  }).finally(() => {
    // Refresh data
    fetchVersions();
    loadOverview();
    progressing.value = false;
  })
}

function uninstallDB() {
  if (!database.value) return;
  
  if (!confirm(`Are you sure you want to uninstall ${database.value.label}? This will remove all data and configurations.`)) {
    return;
  }
  progressing.value = true;
  useNuxtApp().$axiosApi.post(`/databases/${database.value.key}/uninstall`).then((res) => {
    snackbar.add({
      type: "success",
      text: `${database.value.label} Uninstalled successfully`,
    });
  }, (err) => {
    snackbar.add({
      type: "error",
      text: `Failed To Uninstall ${database.value.label}`,
    });
    console.error('Uninstall error:', err);
  }).finally(() => {
    // Refresh data
    fetchVersions();
    loadOverview();
    progressing.value = false;
  })
}

function uninstallSystem() {
  progressing.value = true;
  if (!hardCheckResults.value
    || !database.value
    || !confirm(`Are you sure you want to uninstall System Version? This will remove all data and configurations.`)
  ) return;

  useNuxtApp().$axiosApi.post(`/databases/${database.value.key}/uninstallSystem`).then((res) => {
    snackbar.add({
      type: "success",
      text: `Unnstalled successfully`,
    });
  }, (err) => {
    snackbar.add({
      type: "error",
      text: `Failed To Uninstall ${hardCheckResults.value.db}`,
    });
    console.error('Uninstall error:', err);
  }).finally(() => {
    // Refresh data
    fetchVersions();
    loadOverview();
    runHardCheck();
    progressing.value = false;
  })
}

function togglePassword() {
  showPassword.value = !showPassword.value;
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text)
    .then(() => {
      snackbar.add({
        type: "success",
        text: `Copied To Clipboard`,
      });
    })
    .catch(err => {
      console.error('Failed to copy:', err);
      snackbar.add({
        type: "error",
        text: `Can't Copy To Clipboard`,
      });
    });
}

async function runHardCheck() {
  if (!database.value) return

  hardCheckLoading.value = true
  hardCheckResults.value = null

  useNuxtApp().$axiosApi.get(`/databases/hardCheck?db=${database.value.key}`).then((res) => {
    hardCheckResults.value = res.data.data || {}
  }, (err) => {
    console.error('Hard check failed:', err)
    snackbar.add({
      type: "error",
      text: "Failed to check system for database versions"
    })
  }).finally(() => {
    hardCheckLoading.value = false
  })
}

// Get database version status - backend now returns clean semantic versions
const getVersionStatus = (dbKey) => {
  const versionInfo = versionsState.value[dbKey];
  if (!versionInfo || versionInfo.trim() === "") {
    return { isInstalled: false, version: "", port: getDefaultPort(dbKey) };
  }

  // Backend returns clean semantic versions (x.x.x format)
  return {
    isInstalled: true,
    version: `v${versionInfo}`, // Add 'v' prefix for display consistency
    port: getDefaultPort(dbKey)
  };
};
// Helper function to get default ports for different databases
const getDefaultPort = (dbKey) => {
  const ports = {
    mysql: '3306',
    mariadb: '3306',
    postgres: '5432',
    mongodb: '27017',
    redis: '6379',
    sqlite: 'N/A'
  };
  return ports[dbKey] || '';
};
</script>

<style scoped>
/* Add any custom styles here */
</style>
