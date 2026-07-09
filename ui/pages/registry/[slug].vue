<template>
  <div class="min-h-screen bg-base-200">
    <div class="max-w-4xl mx-auto p-6">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="flex items-center justify-center gap-3 mb-4">
          <Icon :name="manifest?.icon || 'mdi:package-variant'" size="48" class="text-primary" />
          <h1 class="text-3xl font-bold text-base-content">{{ requirements.app || slug }} Installation</h1>
        </div>
        <p class="text-base-content/60">Set up your application with automatic dependency management</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center h-64">
        <Icon name="mdi:loading" class="animate-spin w-8 h-8 text-primary" />
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="alert alert-error mb-6">
        <Icon name="mdi:alert-circle" class="w-5 h-5 shrink-0" />
        <span>{{ error }}</span>
      </div>

      <!-- Step 1: Requirements Check -->
      <div v-else-if="step === 1" class="card p-8">
        <div class="mb-6">
          <h2 class="text-xl font-semibold text-base-content mb-2">System Requirements</h2>
          <p class="text-base-content/60">Checking and installing required dependencies for {{ requirements.app || slug }}</p>
        </div>

        <!-- Requirements Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          <!-- Runtime Requirements -->
          <RuntimeReqPill 
            v-for="rt in requirements.runtimes || []"
            :key="rt.name"
            :runtime="rt"
            @installed="onRuntimeInstalled"
          />
          
          <!-- In-Memory Database Requirements (e.g. Redis) -->
          <RuntimeReqPill 
            v-for="db in requirements.in_memory || []"
            :key="db.name"
            :runtime="db"
            @installed="onDatabaseInstalled"
          />
          
          <!-- Database Requirement -->
          <DBReqPill 
            v-if="requirements.database"
            :database="requirements.database"
            @installed="onDatabaseInstalled"
            @db-selected="onDatabaseSelected"
          />
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            @click="autoInstallDependencies"
            :disabled="installing || allRequirementsMet"
            class="btn btn-primary"
          >
            <Icon v-if="installing" name="mdi:loading" class="animate-spin mr-2" />
            <Icon v-else name="mdi:download" class="mr-2" />
            {{ installing ? 'Installing...' : 'Auto Install Dependencies' }}
          </button>
          
          <button
            @click="goToConfiguration"
            :disabled="!allRequirementsMet"
            class="btn btn-secondary"
          >
            <Icon name="mdi:cog" class="mr-2" />
            Configure Installation
          </button>
        </div>
      </div>

      <!-- Step 2: Application Configuration -->
      <div v-else-if="step === 2" class="card p-8">
        <div class="mb-6">
          <h2 class="text-xl font-semibold text-base-content mb-2">Application Configuration</h2>
          <p class="text-base-content/60">Configure your {{ requirements.app || slug }} installation details</p>
        </div>

        <form @submit.prevent="goToInstallation" class="space-y-6">
          <!-- Deployment Method Selection -->
          <ApplicationsInstallMethods
            v-model="installForm.deploymentMethod"
            v-model:git-repo="installForm.gitRepo"
            v-model:git-branch="installForm.gitBranch"
            v-model:upload-file="installForm.uploadFile"
            v-model:manual-path="installForm.manualPath"
            :slug="slug"
            :deployment-config="deploymentConfig"
          />

          <!-- Installation Information -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <NativeTextField
              v-model="installForm.path"
              label="Installation Path"
              placeholder="/var/www"
              required
            />

            <NativeTextField
              v-model="installForm.domain"
              label="Domain"
              placeholder="example.com"
              required
            />

            <NativeTextField
              v-model="installForm.appName"
              label="Application Name"
              placeholder="My App"
              required
            />
          </div>

          <template v-if="requirements.variables && requirements.variables.length">
            <hr class="my-6 border-base-300" />
            
            <div class="mb-4">
              <h3 class="text-lg font-medium text-base-content mb-2">Application Variables</h3>
              <p class="text-base-content/60 text-sm">Configure required variables for {{ requirements.app || slug }}</p>
            </div>
            
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div v-for="(variable, index) in requirements.variables" :key="index" class="w-full">
                <NativeTextField
                  v-model="installForm.vars[variable.Key || variable.key]"
                  :label="variable.Prompt || variable.prompt"
                  :type="(variable.Secret || variable.secret) ? 'password' : 'text'"
                  :required="variable.Required || variable.required"
                  :placeholder="(variable.Default || variable.default) || ''"
                  :hint="(variable.Default || variable.default) ? `Default: ${variable.Default || variable.default}` : ''"
                />
              </div>
            </div>
          </template>

          <!-- Navigation Buttons -->
          <div class="flex justify-between pt-6">
            <button
              type="button"
              @click="step = 1"
              class="btn btn-ghost"
            >
              <Icon name="mdi:arrow-left" class="mr-2" />
              Back to Requirements
            </button>
            
            <button
              type="submit"
              class="btn btn-primary"
            >
              Continue to Installation
              <Icon name="mdi:arrow-right" class="ml-2" />
            </button>
          </div>
        </form>
      </div>

      <!-- Step 3: Installation -->
      <div v-else-if="step === 3" class="card p-8">
        <div class="mb-6">
          <h2 class="text-xl font-semibold text-base-content mb-2">Installation Summary</h2>
          <p class="text-base-content/60">Review your configuration and install {{ requirements.app || slug }}</p>
        </div>

        <!-- Configuration Summary -->
        <div class="bg-base-200 rounded-lg p-6 mb-6">
          <h3 class="font-semibold text-base-content mb-4">Installation Details</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <span class="font-medium">Application:</span> {{ requirements.app || slug }}
            </div>
            <div>
              <span class="font-medium">Path:</span> {{ installForm.path }}
            </div>
            <div>
              <span class="font-medium">Domain:</span> {{ installForm.domain }}
            </div>
            <div>
              <span class="font-medium">Name:</span> {{ installForm.appName }}
            </div>
          </div>
          
          <!-- Variables Summary -->
          <div v-if="Object.keys(installForm.vars).length" class="mt-4">
            <h4 class="font-medium text-base-content mb-2">Variables:</h4>
            <div class="space-y-1 text-sm">
              <div v-for="(value, key) in installForm.vars" :key="key" class="flex justify-between">
                <span class="font-medium">{{ key }}:</span>
                <span>{{ value ? '••••••' : 'Not set' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Installation Button -->
        <div class="flex justify-between">
          <button
            @click="step = 2"
            class="btn btn-ghost"
          >
            <Icon name="mdi:arrow-left" class="mr-2" />
            Back to Configuration
          </button>
          
          <button
            @click="installApplication"
            :disabled="installing"
            class="btn btn-primary btn-lg"
          >
            <Icon v-if="installing" name="mdi:loading" class="animate-spin mr-2" />
            <Icon v-else name="mdi:download" class="mr-2" />
            {{ installing ? `Installing ${requirements.app || slug}...` : `Install ${requirements.app || slug}` }}
          </button>
        </div>
      </div>

      <!-- Installation Status Modal -->
      <NativeDialog v-model="installStatus" v-if="installStatus">
        <div class="card bg-base-100 shadow-xl border border-base-300 p-6 max-w-2xl w-full mx-4">
          <div class="card-body">
            <h3 class="card-title text-lg">Installation Status</h3>
            <div :class="installStatus.success ? 'alert alert-success' : 'alert alert-error'">
              <div class="flex items-center">
                <Icon :name="installStatus.success ? 'mdi:check-circle' : 'mdi:close-circle'" class="w-5 h-5 mr-2" />
                <span>
                  {{ installStatus.success ? 'Success' : 'Failed' }}
                </span>
              </div>
              <div>
                <p class="text-sm">{{ installStatus.message }}</p>
              </div>
            </div>
            
            <!-- Installation Steps -->
            <div v-if="installStatus.steps && installStatus.steps.length" class="mt-4">
              <h4 class="text-sm font-medium text-base-content/80 mb-3">Installation Steps:</h4>
              <div class="space-y-3 max-h-64 overflow-y-auto">
                <div v-for="(step, index) in installStatus.steps" :key="index" class="border border-base-300 rounded-lg p-3 bg-base-200">
                  <div class="flex items-start justify-between mb-2">
                    <div class="flex items-center">
                      <div class="badge badge-primary badge-sm mr-2">{{ step.Index + 1 }}</div>
                      <h5 class="font-medium text-sm text-base-content">{{ step.Name }}</h5>
                    </div>
                    <div class="flex items-center">
                      <Icon 
                        v-if="step.Error" 
                        name="mdi:alert-circle" 
                        class="w-4 h-4 text-error" 
                      />
                      <Icon 
                        v-else-if="step.Output" 
                        name="mdi:check-circle" 
                        class="w-4 h-4 text-success" 
                      />
                      <Icon 
                        v-else 
                        name="mdi:minus-circle" 
                        class="w-4 h-4 text-base-content/50" 
                      />
                    </div>
                  </div>
                  
                  <!-- Step Output -->
                  <div v-if="step.Output" class="mt-2">
                    <div class="text-xs font-medium text-base-content/60 mb-1">Output:</div>
                    <div class="bg-base-100 rounded p-2 text-xs font-mono text-base-content/80 whitespace-pre-wrap">{{ step.Output }}</div>
                  </div>
                  
                  <!-- Step Error -->
                  <div v-if="step.Error" class="mt-2">
                    <div class="text-xs font-medium text-error/60 mb-1">Error:</div>
                    <div class="bg-error/10 border border-error/20 rounded p-2 text-xs font-mono text-error whitespace-pre-wrap">{{ step.Error }}</div>
                  </div>
                </div>
              </div>
            </div>

            <div class="card-actions justify-end mt-4">
              <button @click="installStatus = null" class="btn btn-sm">Close</button>
            </div>
          </div>
        </div>
      </NativeDialog>
    </div>
  </div>
</template>

<script setup>
const route = useRoute();
const slug = route.params.slug;

// Reactive state
const loading = ref(true);
const error = ref('');
const requirements = ref({});
const appManifest = ref({});
const installing = ref(false);
const installStatus = ref(null);
const step = ref(1);
const deploymentConfig = ref({});

// Form data
const installForm = ref({
  path: '/var/www',
  domain: '',
  appName: '',
  vars: {},
  deploymentMethod: 'git',
  gitRepo: '',
  gitBranch: '',
  uploadFile: null,
  manualPath: ''
});

// Computed properties
const allRequirementsMet = computed(() => {
  const runtimesMet = !requirements.value.runtimes || 
    requirements.value.runtimes.every(rt => rt.available);
  const inMemoryMet = !requirements.value.in_memory ||
    requirements.value.in_memory.every(db => db.available);
  const databaseMet = !requirements.value.database || 
    (requirements.value.database.selected && 
     requirements.value.database.options?.some(opt => opt.name === requirements.value.database.selected && opt.available));
  return runtimesMet && inMemoryMet && databaseMet;
});

onMounted(() => {
  loadRequirements();
});

const loadRequirements = async () => {
  try {
    loading.value = true;
    error.value = '';
    
    // Fetch requirements
    const reqResponse = await useNuxtApp().$axiosApi.get(`/registry/${slug}/requirements`);
    if (reqResponse.data.success) {
      requirements.value = reqResponse.data;
      
      // Set deployment config from requirements response
      deploymentConfig.value = reqResponse.data.deployment || {};
      
      // Set default deployment method
      if (deploymentConfig.value.default) {
        installForm.value.deploymentMethod = deploymentConfig.value.default;
      }
      
      // Set default git values if available
      if (deploymentConfig.value.git) {
        installForm.value.gitRepo = deploymentConfig.value.git.source || '';
        installForm.value.gitBranch = deploymentConfig.value.git.branch || '';
      }
    } else {
      throw new Error(reqResponse.data.error || 'Failed to load requirements');
    }

    // Set default values
    installForm.value.appName = requirements.value.app || slug;
    installForm.value.domain = `${slug}.localhost`;

  } catch (err) {
    error.value = err.response?.data?.error || err.message || 'Failed to load application details';
  } finally {
    loading.value = false;
  }
};

const autoInstallDependencies = async () => {
  // For now, just show a message that this feature is coming
  // In a real implementation, this would call an API to install dependencies
  alert('Auto-install feature coming soon! Please install dependencies manually and refresh the page.');
};

const getRequirementIcon = (name) => {
  const icons = {
    'php': 'mdi:php',
    'node': 'mdi:nodejs',
    'python': 'mdi:python',
    'golang': 'mdi:language-go',
    'database': 'mdi:database',
    'mysql': 'mdi:mysql',
    'mariadb': 'mdi:database',
    'postgresql': 'mdi:postgresql',
    'nginx': 'mdi:nginx',
    'caddy': 'mdi:web',
    'composer': 'mdi:composer',
    'npm': 'mdi:npm',
    'yarn': 'mdi:yarn',
    'git': 'mdi:git',
  };
  return icons[name.toLowerCase()] || 'mdi:package-variant';
};

const goToConfiguration = () => {
  if (allRequirementsMet.value) {
    step.value = 2;
  }
};

const goToInstallation = () => {
  if (allRequirementsMet.value) {
    step.value = 3;
  }
};

const onRuntimeInstalled = async (data) => {
  try {
    // Refresh requirements after installation
    await loadRequirements();
    
    // Show success message
    useToaster(`${data.name} installed successfully`, "bg-green-500 text-white");
  } catch (err) {
    useToaster(`Failed to install ${data.name}`, "bg-red-500 text-white");
  }
};

const onDatabaseInstalled = async (data) => {
  try {
    // Refresh requirements after installation
    await loadRequirements();
    
    // Show success message
    useToaster(`${data.name} installed successfully`, "bg-green-500 text-white");
  } catch (err) {
    useToaster(`Failed to install ${data.name}`, "bg-red-500 text-white");
  }
};

const onDatabaseSelected = (dbName) => {
  // Update the database selection
  if (requirements.value.database) {
    requirements.value.database.selected = dbName;
  }
};

const installApplication = async () => {
  try {
    installing.value = true;
    installStatus.value = null;

    const payload = {
      ...installForm.value,
      databaseEngine: requirements.value.database?.selected || ''
    };

    const response = await useNuxtApp().$axiosApi.post(`/registry/${slug}/install`, payload);
    
    if (response.data.success) {
      installStatus.value = {
        success: true,
        message: response.data.message || 'Application installed successfully',
        steps: response.data.steps || []
      };
      
      // Reset form and go back to step 1
      installForm.value = {
        path: '/var/www',
        domain: '',
        appName: '',
        vars: {},
        deploymentMethod: 'git',
        gitRepo: '',
        gitBranch: '',
        uploadFile: null,
        manualPath: ''
      };
      step.value = 1;
    } else {
      installStatus.value = {
        success: false,
        message: response.data.error || 'Installation failed',
        steps: response.data.steps || []
      };
    }
  } catch (err) {
    installStatus.value = {
      success: false,
      message: err.response?.data?.error || err.message || 'Installation failed',
      steps: err.response?.data?.steps || []
    };
  } finally {
    installing.value = false;
  }
};
</script>