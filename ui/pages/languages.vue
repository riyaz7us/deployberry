<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
    <div class="space-y-2">
      <h1 class="text-xl font-semibold">Programming Languages</h1>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <LangPill v-for="langKey in languages" :key="langKey" :language-key="langKey" @click="selectLanguage({key: langKey})"
          :version="getVersionStatus(langKey).version" 
          :installed="getVersionStatus(langKey).isInstalled"
          :loadingOverview="loadingOverview" />
      </div>
    </div>

    <NativeDialog v-model="language">
      <NativeInstalling v-if="progressing" />
      <div v-else class="space-y-4">
        <h2 class="text-base-content font-semibold">Manage {{ language?.label }}</h2>
        <div v-if="language" class="space-y-3">
          <div class="flex items-center gap-3">
            <span class="text-base-content/70">Available {{ language.label }} versions</span>
            <button @click="fetchVersions"
              class="text-xs px-2 py-1 border border-base-300 rounded hover:bg-base-300 transition-colors"
              :disabled="loadingVersions">Refresh</button>
            <button @click="runHardCheck"
              class="text-xs px-2 py-1 border border-primary text-primary rounded hover:bg-primary/30 transition-colors"
              :disabled="hardCheckLoading">
              <span v-if="!hardCheckLoading">Check System</span>
              <span v-else class="flex items-center">
                <Icon name="mdi:loading" size="14" class="animate-spin mr-1" /> Checking...
              </span>
            </button>
          </div>

          <div v-if="hardCheckResults" class="mt-4 p-3 card">
            <h3 class="text-sm font-bold text-base-content/70 mb-2">System Detected Versions</h3>
            <p class="text-sm font-medium mb-2 italic text-red-500 max-w-3xl">
              Following are the system detected versions of your language. If they do not match with the versions
              available below, you might need to uninstall them. If you have important data or configurations, please
              backup before proceeding.
            </p>
            <button @click="uninstallSystem()" class="text-xs px-2 py-1 rounded btn btn-warning"
              :disabled="loadingVersions">Uninstall</button>
            <div class="space-y-1">
              <div v-for="(version, key) in hardCheckResults" :key="key" class="flex justify-between text-sm">
                <span class="text-base-content/60">{{ key }}:</span>
                <span class="font-mono text-base-content/80">{{ version || "Not found" }}</span>
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

            <div v-else-if="versionEntries.length === 0" class="text-center py-4 text-base-content/60">No versions
              available</div>
            <div v-for="e in versionEntries" :key="e.version" class="border border-base-300 rounded p-3 bg-base-200">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-2">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-medium">{{ language.label }} {{ e.version }}</span>
                  <span v-if="e.active"
                    class="text-xs px-2 py-0.5 rounded bg-success text-success-content font-medium">Active</span>
                  <span v-else-if="e.installed"
                    class="text-xs px-2 py-0.5 rounded bg-info text-info-content font-medium">Installed</span>
                </div>
                <div class="flex gap-2 justify-end w-full sm:w-auto border-t border-base-300/30 pt-2 sm:pt-0 sm:border-t-0 flex-wrap">
                  <button v-if="e.installed && !e.active" @click="activateLanguage(e.version)"
                    class="text-xs px-2 py-1 rounded btn btn-primary btn-sm" :disabled="loadingVersions">
                    Activate
                  </button>
                  <button v-else-if="e.installed && e.active" @click="deactivateLanguage()"
                    class="text-xs px-2 py-1 rounded btn btn-error btn-sm" :disabled="loadingVersions">
                    Deactivate
                  </button>
                  <button v-if="e.installed && e.active" @click="navigateTo(`/tool-configs/${language.key}`)"
                    class="text-xs px-2 py-1 rounded btn btn-secondary btn-sm">Configure</button>
                  <button v-if="e.installed" @click="uninstallLanguage(e.version)"
                    class="text-xs px-2 py-1 rounded btn btn-warning btn-sm" :disabled="loadingVersions">
                    Uninstall
                  </button>
                  <button v-else-if="!e.installed" @click="installLanguage(e.version)"
                    class="text-xs px-2 py-1 rounded btn btn-neutral btn-sm" :disabled="loadingVersions">
                    Install
                  </button>
                </div>
              </div>

              <div v-if="e.installed && e.extensions && e.extensions.length > 0"
                class="mt-2 text-sm text-base-content/70">
                <div class="flex items-start justify-between">
                  <span>Extensions:</span>
                  <div class="flex flex-wrap gap-1 ml-2">
                    <span v-for="ext in e.extensions" :key="ext" class="text-xs px-1 py-0.5 bg-base-300 rounded">{{ ext
                    }}</span>
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
const snackbar = useSnackbar();

const versionsState = ref({});

const languages = [
  "php", "node", "python", "golang"
];

const dialog = ref(false);
const language = ref(null);
const versionEntries = ref([]);
const warnings = ref("");
const loadingOverview = ref(false);
const loadingVersions = ref(false);
const progressing = ref(false);
const hardCheckResults = ref(null);
const hardCheckLoading = ref(false);

onMounted(() => {
  loadOverview();
});

function loadOverview() {
  loadingOverview.value = true;
  useNuxtApp()
    .$axiosApi.get("/languages/")
    .then((res) => {
      // Ensure we have an object even if the response is empty
      versionsState.value = res.data?.data || {};

      // Log for debugging
      console.log("Loaded language versions:", versionsState.value);
    })
    .catch((error) => {
      snackbar.add({
        type: "error",
        text: `Failed To Load Language Status`,
      });
    })
    .finally(() => {
      loadingOverview.value = false;
    });
}

function selectLanguage(langData) {
  // Create full language object for dialog usage
  const langKey = langData.key;
  language.value = {
    key: langKey,
    label: langKey.charAt(0).toUpperCase() + langKey.slice(1) // Simple fallback
  };
  versionEntries.value = [];
  hardCheckResults.value = null;
  dialog.value = true;
  fetchVersions();
}

function fetchVersions() {
  if (!language.value) return;

  loadingVersions.value = true;
  useNuxtApp()
    .$axiosApi.get(`/languages/${language.value.key}/versions`)
    .then((res) => {
      versionEntries.value = Array.isArray(res.data.data) ? res.data.data : [];
      warnings.value = res.data.warnings;
    })
    .catch((error) => {
      snackbar.add({
        type: "error",
        text: `Failed To Fetch Language Versions`,
      });
      versionEntries.value = [];
    })
    .finally(() => {
      loadingVersions.value = false;
    });
}

async function installLanguage(version) {
  if (!language.value) return;
  console.log("versionEntries", versionEntries.value);
  progressing.value = true;
  useNuxtApp()
    .$axiosApi.post(`/languages/${language.value.key}/install?version=${version}`)
    .then(
      (res) => {
        loadOverview();
        fetchVersions();
        snackbar.add({
          type: "success",
          text: `${language.value.label} ${version} Installed successfully`,
        });
      },
      (err) => {
        snackbar.add({
          type: "error",
          text: `Failed To Install ${language.value.label} ${version}`,
        });
        console.error("Installation error:", err);
      }
    )
    .finally(() => {
      progressing.value = false;
    });
}

function activateLanguage(v) {
  progressing.value = true;
  if (!language.value) return;
  useNuxtApp()
    .$axiosApi.post(`/languages/${language.value.key}/activate`, { version: v })
    .then(
      (res) => {
        snackbar.add({
          type: "success",
          text: `${language.value.label} Activated successfully`,
        });
      },
      (err) => {
        snackbar.add({
          type: "error",
          text: `Failed To Activate ${language.value.label}`,
        });
        console.error("Activation error:", err);
      }
    )
    .finally(() => {
      // Refresh data
      fetchVersions();
      loadOverview();
      progressing.value = false;
    });
}

function uninstallLanguage(version) {
  if (!language.value) return;

  if (!confirm(`Are you sure you want to uninstall ${language.value.label} ${version}? This will remove all data and configurations.`)) {
    return;
  }
  progressing.value = true;
  useNuxtApp()
    .$axiosApi.post(`/languages/${language.value.key}/uninstall?version=${version}`)
    .then(
      (res) => {
        snackbar.add({
          type: "success",
          text: `${language.value.label} ${version} Uninstalled successfully`,
        });
      },
      (err) => {
        snackbar.add({
          type: "error",
          text: `Failed To Uninstall ${language.value.label}`,
        });
        console.error("Uninstall error:", err);
      }
    )
    .finally(() => {
      // Refresh data
      fetchVersions();
      loadOverview();
      progressing.value = false;
    });
}

function uninstallSystem() {
  progressing.value = true;
  if (!hardCheckResults.value || !language.value || !confirm(`Are you sure you want to uninstall System Version? This will remove all data and configurations.`)) return;

  useNuxtApp()
    .$axiosApi.post(`/languages/${language.value.key}/uninstallSystem`)
    .then(
      (res) => {
        snackbar.add({
          type: "success",
          text: `Uninstalled successfully`,
        });
      },
      (err) => {
        snackbar.add({
          type: "error",
          text: `Failed To Uninstall ${hardCheckResults.value.language}`,
        });
        console.error("Uninstall error:", err);
      }
    )
    .finally(() => {
      // Refresh data
      fetchVersions();
      loadOverview();
      runHardCheck();
      progressing.value = false;
    });
}

async function runHardCheck() {
  if (!language.value) return;

  hardCheckLoading.value = true;
  hardCheckResults.value = null;

  useNuxtApp()
    .$axiosApi.get(`/languages/hardCheck?language=${language.value.key}`)
    .then(
      (res) => {
        hardCheckResults.value = res.data.data || {};
      },
      (err) => {
        console.error("Hard check failed:", err);
        snackbar.add({
          type: "error",
          text: "Failed to check system for language versions",
        });
      }
    )
    .finally(() => {
      hardCheckLoading.value = false;
    });
}

// Get version status - backend now returns clean semantic versions
const getVersionStatus = (langKey) => {
  const versionInfo = versionsState.value[langKey];
  if (!versionInfo || versionInfo.trim() === "") {
    return { isInstalled: false, version: "" };
  }

  // Backend returns clean semantic versions (x.x.x format)
  return {
    isInstalled: true,
    version: `v${versionInfo}`, // Add 'v' prefix for display consistency
  };
};

function deactivateLanguage() {
  if (!language.value) return;
  progressing.value = true;

  useNuxtApp()
    .$axiosApi.post(`/languages/${language.value.key}/deactivate`)
    .then(
      (res) => {
        snackbar.add({
          type: "success",
          text: `${language.value.label} Deactivated Successfully`,
        });
      },
      (err) => {
        snackbar.add({
          type: "error",
          text: `Failed To Deactivate ${language.value.label}`,
        });
        console.error("Deactivation error:", err);
      }
    )
    .finally(() => {
      // Refresh data
      fetchVersions();
      loadOverview();
      progressing.value = false;
    });
}

function openPHPConfig() {
  // Navigate to PHP configuration page
  navigateTo("/tool-configs/php");
}
</script>

<style scoped></style>
