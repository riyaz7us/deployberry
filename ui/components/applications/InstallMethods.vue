<template>
  <div v-if="deploymentMethods.length > 0" class="space-y-6">
    <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-box">
      <input type="checkbox" v-model="isExpanded" /> 
      <div class="collapse-title text-lg font-medium">
        Deployment Configuration
        <span class="text-sm font-normal text-base-content/60 block mt-1">Configure source code deployment options</span>
      </div>
      <div class="collapse-content">
        <div class="space-y-6 pt-4 border-t border-base-200">
          <!-- Installation Method Selection -->
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-3">
              <h3 class="text-lg font-medium text-base-content mb-4">Deployment Method</h3>
              <NativeBoxSelector
                v-model="selectedMethod"
                :options="deploymentMethods"
              />
            </div>
          </div>

    <!-- Git Installation Fields -->
    <div v-if="selectedMethod === 'git'" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="lg:col-span-2">
        <h3 class="text-lg font-medium text-base-content mb-4">Git Repository Configuration</h3>
        <div class="bg-base-200/50 border border-base-300 rounded-lg p-4">
          <p class="text-base-content/60 text-sm mb-4">
            Enter your Git repository URL to clone your application.
            <SSHKeyManager />
          </p>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <NativeTextField
                v-model="gitRepo"
                label="Git Repository URL"
                :placeholder="deploymentConfig.git?.source || 'https://github.com/username/my-app.git'"
                icon="mdi:git"
                :monospace="true"
                required
                helper-text="Ensure that you have added the SSH key to your Git server for private repositories."
              />
            </div>

            <NativeTextField
              v-model="gitBranch"
              label="Branch (Optional)"
              :placeholder="deploymentConfig.git?.branch || 'main'"
              icon="mdi:source-branch"
              :monospace="true"
              helper-text="Leave empty to use default branch"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Upload Installation Fields -->
    <div v-if="selectedMethod === 'upload'" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="lg:col-span-2">
        <h3 class="text-lg font-medium text-base-content mb-4">Upload Application</h3>
        <div class="bg-base-200/50 border border-base-300 rounded-lg p-4">
          <p class="text-base-content/60 text-sm mb-4">
            Upload your application files or a zip archive.
          </p>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <NativeImageUpload
                v-model="uploadFile"
                label="Application Files"
                accept=".zip,.tar.gz"
                icon="mdi:upload"
                required
                helper-text="Select a zip file or tar.gz archive containing your application"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Manual Path Fields -->
    <div v-if="selectedMethod === 'manual'" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="lg:col-span-2">
        <h3 class="text-lg font-medium text-base-content mb-4">Manual Path</h3>
        <div class="bg-base-200/50 border border-base-300 rounded-lg p-4">
          <p class="text-base-content/60 text-sm mb-4">
            Use an existing project from a local path on the server.
          </p>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <NativeTextField
                v-model="manualPath"
                label="Project Path"
                placeholder="/home/user/my-project"
                icon="mdi:folder-open"
                :monospace="true"
                required
                helper-text="Full path to your existing project directory."
              />
            </div>
          </div>
        </div>
      </div>
    </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  slug: {
    type: String,
    required: true
  },
  deploymentConfig: {
    type: Object,
    default: () => ({})
  }
})

const selectedMethod = defineModel({ default: 'git' })
const gitRepo = defineModel('gitRepo', { default: '' })
const gitBranch = defineModel('gitBranch', { default: '' })
const uploadFile = defineModel('uploadFile', { default: null })
const manualPath = defineModel('manualPath', { default: '' })

// Set default method from deployment config
if (props.deploymentConfig.default) {
  selectedMethod.value = props.deploymentConfig.default
}

const isExpanded = ref(true);

// All possible deployment methods
const ALL_METHODS = [
  {
    value: "git",
    label: "Git Repository",
    icon: "mdi:git",
    iconColor: "text-primary",
    description: "Clone from Git repository"
  },
  {
    value: "upload",
    label: "Upload Files",
    icon: "mdi:upload",
    iconColor: "text-success",
    description: "Upload files or archive"
  },
  {
    value: "manual",
    label: "Manual Path",
    icon: "mdi:folder-open",
    iconColor: "text-info",
    description: "Use existing local path"
  }
];

// Fresh Install option
const NONE_METHOD = {
  value: "none",
  label: "Fresh Install (Skip Deployment)",
  icon: "mdi:sparkles",
  iconColor: "text-warning",
  description: "Install from scratch without cloning/uploading files"
};

// Deployment methods options
const deploymentMethods = ref([...ALL_METHODS]);

// Load deployment methods from manifest
const loadDeploymentMethods = (deploymentConfig) => {
  try {
    // Filter available methods based on manifest
    const available = deploymentConfig?.available || {}
    let filtered = ALL_METHODS.filter(method => available[method.value])
    
    // Add "Fresh Install" option if deployment is not strictly required
    if (deploymentConfig?.required === false) {
      filtered = [NONE_METHOD, ...filtered];
    }
    
    // Update methods
    deploymentMethods.value = filtered
    
    // Set default method if available
    if (deploymentConfig?.required === false) {
      selectedMethod.value = 'none';
      isExpanded.value = false;
    } else if (deploymentConfig?.default && available[deploymentConfig.default]) {
      selectedMethod.value = deploymentConfig.default;
      isExpanded.value = true;
    } else if (filtered.length > 0) {
      selectedMethod.value = filtered[0].value;
      isExpanded.value = true;
    } else {
      selectedMethod.value = 'none';
      isExpanded.value = false;
    }
    
  } catch (error) {
    console.error('Failed to load deployment methods:', error)
    // Keep default methods if API fails
    selectedMethod.value = 'none'
    deploymentMethods.value = []
  }
}

// Watch for deployment config changes
watch(() => props.deploymentConfig, (newConfig) => {
  if (newConfig) {
    loadDeploymentMethods(newConfig)
  }
}, { immediate: true })
</script>
