<template>
  <div class="min-h-screen bg-base-200">
    <div class="max-w-4xl mx-auto p-4 sm:p-6">
      <!-- Header -->
      <div class="text-center mb-8 border-b border-base-300 pb-6">
        <div class="flex items-center justify-center gap-3 mb-3">
          <Icon :name="slug === 'custom' ? 'mdi:docker' : (requirements.icon || 'mdi:package-variant')" size="48"
            class="text-primary" />
          <h1 class="text-3xl font-bold text-base-content">{{ slug === 'custom' ? 'Deploy Custom Container' :
            (requirements.app || slug) }}</h1>
        </div>
        <p class="text-base-content/60 text-sm">Configure and launch your containerized service using Podman Compose
          orchestration</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex flex-col justify-center items-center h-64 gap-3">
        <span class="loading loading-spinner loading-lg text-primary"></span>
        <span class="text-xs text-base-content/50">Fetching requirements and manifest settings...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="alert alert-error mb-6">
        <Icon name="mdi:alert-circle" class="w-5 h-5 shrink-0" />
        <span>{{ error }}</span>
      </div>

      <template v-else>
        <!-- Dependency Warning -->
        <div v-if="!requirements.podman_installed" class="alert alert-warning mb-6 flex items-start gap-3">
          <Icon name="mdi:alert" class="w-6 h-6 shrink-0 mt-0.5" />
          <div>
            <h3 class="font-bold text-sm">Podman Compose Not Installed</h3>
            <p class="text-xs mt-0.5">Podman and podman-compose binaries were not found on the host
              system. The panel will attempt to automatically install them during the application build phase.</p>
          </div>
        </div>

        <!-- Step 1: Configuration Form -->
        <div v-if="step === 1" class="card bg-base-100 shadow-xl border border-base-300 p-6 sm:p-8 space-y-6">
          <div>
            <h2 class="text-xl font-bold text-base-content">Deployment Options</h2>
            <p class="text-xs text-base-content/60 mt-1">Setup the source, routing, database, and custom environment
              parameters.</p>
          </div>

          <form @submit.prevent="goToSummary" class="space-y-6">
            <!-- Custom UI deployment selector (Only for slug == 'custom') -->
            <div v-if="slug === 'custom'" class="space-y-4 bg-base-200 p-4 rounded-xl border border-base-300">
              <h3 class="text-sm font-semibold text-base-content">Container Source Type</h3>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
                <button type="button" v-for="mode in ['image', 'compose', 'git']" :key="mode"
                  @click="customMode = mode"
                  :class="['btn btn-sm capitalize', customMode === mode ? 'btn-primary' : 'btn-ghost bg-base-100']">
                  {{ mode }}
                </button>
              </div>

              <!-- Mode-specific Inputs -->
              <div class="space-y-4 pt-3 border-t border-base-300">
                <!-- Image Mode -->
                <div v-if="customMode === 'image'" class="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div class="md:col-span-2">
                    <label class="label text-xs font-semibold">Image Name</label>
                    <input v-model="installForm.image" required type="text" class="input input-sm input-bordered w-full"
                      placeholder="e.g. docker.io/library/nginx:alpine" />
                  </div>
                  <div>
                    <label class="label text-xs font-semibold">Container Port</label>
                    <input v-model.number="installForm.container_port" required type="number"
                      class="input input-sm input-bordered w-full" placeholder="e.g. 80" />
                  </div>
                </div>

                <!-- Compose Mode -->
                <div v-if="customMode === 'compose'" class="space-y-2">
                  <label class="label text-xs font-semibold">Docker Compose Template</label>
                  <p class="text-[10px] text-base-content/50">Must expose web service to port <code
                      class="text-primary">{HOST_PORT}</code> (e.g. ports: ["{HOST_PORT}:80"]).</p>
                  <textarea v-model="installForm.compose_template" required
                    class="textarea textarea-bordered w-full font-mono text-xs h-48"
                    placeholder="version: '3.8'&#10;services:&#10;  app:&#10;    image: nginx:alpine&#10;    ports:&#10;      - &quot;{HOST_PORT}:80&quot;"></textarea>
                </div>

                <!-- Git/Repo Mode -->
                <div v-if="customMode === 'git'" class="space-y-4">
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div class="md:col-span-2">
                      <label class="label text-xs font-semibold">Git Repository URL</label>
                      <input v-model="installForm.gitRepo" required type="text"
                        class="input input-sm input-bordered w-full"
                        placeholder="https://github.com/username/repo.git" />
                    </div>
                    <div>
                      <label class="label text-xs font-semibold">Branch (Optional)</label>
                      <input v-model="installForm.gitBranch" type="text" class="input input-sm input-bordered w-full"
                        placeholder="main" />
                    </div>
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label class="label text-xs font-semibold">Container Port</label>
                      <input v-model.number="installForm.container_port" required type="number"
                        class="input input-sm input-bordered w-full" placeholder="e.g. 80" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Routing & Paths Configuration -->
            <div class="space-y-4">
              <h3 class="text-sm font-semibold text-base-content border-b border-base-200 pb-2">Location & Domain
                Mapping</h3>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label class="label text-xs font-semibold">App Title / Slug</label>
                  <input v-model="installForm.appName" required type="text" class="input input-bordered w-full"
                    placeholder="my-blog" />
                </div>
                <div class="md:col-span-2">
                  <label class="label text-xs font-semibold">Proxy Domain</label>
                  <input v-model="installForm.domain" required type="text" class="input input-bordered w-full"
                    placeholder="blog.example.com" />
                </div>
              </div>
              <div>
                <label class="label text-xs font-semibold">Host Installation Directory</label>
                <input v-model="installForm.path" required type="text" class="input input-bordered w-full"
                  placeholder="/opt/panel17/apps" />
              </div>
            </div>



            <!-- Application Variables -->
            <div v-if="requirements.variables && requirements.variables.length" class="space-y-4">
              <h3 class="text-sm font-semibold text-base-content border-b border-base-200 pb-2">Environment
                Configuration</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div v-for="variable in requirements.variables" :key="variable.key">
                  <label class="label text-xs font-semibold">
                    <span>{{ variable.prompt }}</span>
                    <span v-if="variable.required" class="text-error text-[10px]">*required</span>
                  </label>

                  <select v-if="variable.type === 'select'" v-model="installForm.vars[variable.key]"
                    class="select select-bordered w-full">
                    <option v-for="opt in variable.options" :key="opt" :value="opt">{{ opt }}</option>
                  </select>

                  <input v-else :type="variable.secret ? 'password' : 'text'" v-model="installForm.vars[variable.key]"
                    :required="variable.required" class="input input-bordered w-full"
                    :placeholder="variable.default || ''" />
                  <p class="text-[10px] text-base-content/50 mt-1">{{ variable.helper }}</p>
                </div>
              </div>
            </div>

            <!-- Custom Host Port (Advanced) -->
            <div class="collapse collapse-arrow border border-base-300 bg-base-100 rounded-xl">
              <input type="checkbox" />
              <div class="collapse-title text-xs font-semibold text-base-content/60">
                Advanced Network Port Settings
              </div>
              <div class="collapse-content space-y-4">
                <div class="max-w-xs">
                  <label class="label text-xs font-semibold">Custom Host Port Mapping</label>
                  <p class="text-[10px] text-base-content/50 mb-2">Leave blank to let the panel auto-allocate a free
                    port (9000-10000).</p>
                  <input v-model="installForm.vars.APP_PORT" type="text" class="input input-sm input-bordered w-full"
                    placeholder="e.g. 9090" />
                </div>
              </div>
            </div>

            <!-- Submit buttons -->
            <div class="flex justify-between border-t border-base-200 pt-6">
              <nuxt-link to="/containerapps" class="btn btn-ghost">
                <Icon name="mdi:arrow-left" class="mr-2" />
                Back to Registry
              </nuxt-link>
              <button type="submit" class="btn btn-primary">
                Continue to Summary
                <Icon name="mdi:arrow-right" class="ml-2" />
              </button>
            </div>
          </form>
        </div>

        <!-- Step 2: Deployment Summary -->
        <div v-else-if="step === 2" class="card bg-base-100 shadow-xl border border-base-300 p-6 sm:p-8 space-y-6">
          <div>
            <h2 class="text-xl font-bold text-base-content">Deployment Review</h2>
            <p class="text-xs text-base-content/60 mt-1">Review the configurations before initiating podman-compose up.
            </p>
          </div>

          <div class="bg-base-200 rounded-xl p-5 space-y-3 text-sm">
            <div class="grid grid-cols-3 gap-2 border-b border-base-300 pb-2">
              <span class="font-semibold text-base-content/60">Application:</span>
              <span class="col-span-2 font-medium text-base-content">{{ slug === 'custom' ? `Custom (${customMode})` :
                requirements.app }}</span>
            </div>
            <div class="grid grid-cols-3 gap-2 border-b border-base-300 pb-2">
              <span class="font-semibold text-base-content/60">Routing Domain:</span>
              <span class="col-span-2 font-mono text-primary font-bold">{{ installForm.domain }}</span>
            </div>
            <div class="grid grid-cols-3 gap-2 border-b border-base-300 pb-2">
              <span class="font-semibold text-base-content/60">Host Directory:</span>
              <span class="col-span-2 font-mono text-base-content">{{ installForm.path }}/{{ installForm.domain
              }}</span>
            </div>

            <div class="grid grid-cols-3 gap-2" v-if="installForm.vars.APP_PORT">
              <span class="font-semibold text-base-content/60">Allocated Host Port:</span>
              <span class="col-span-2 font-mono text-base-content">{{ installForm.vars.APP_PORT }}</span>
            </div>
          </div>

          <div class="flex justify-between">
            <button @click="step = 1" class="btn btn-ghost">
              <Icon name="mdi:arrow-left" class="mr-2" />
              Edit Settings
            </button>
            <button @click="launchDeployment" :disabled="deploying" class="btn btn-primary btn-md sm:btn-lg">
              <span v-if="deploying" class="loading loading-spinner mr-2"></span>
              <Icon v-else name="mdi:cloud-upload" class="mr-2" />
              {{ deploying ? 'Deploying...' : 'Deploy Container' }}
            </button>
          </div>
        </div>

        <!-- Progress Dialog/Modal -->
        <dialog ref="progressModal" class="modal bg-base-300/60">
          <div class="modal-box max-w-2xl border border-base-300 shadow-2xl p-6">
            <h3 class="font-bold text-lg text-base-content border-b border-base-300 pb-3 flex items-center gap-2">
              <span v-if="deployStatus === 'installing'" class="loading loading-spinner text-primary"></span>
              <Icon v-else-if="deployStatus === 'success'" name="mdi:check-circle" class="text-success" size="24" />
              <Icon v-else name="mdi:alert-circle" class="text-error" size="24" />
              Deployment Console
            </h3>

            <p class="py-4 text-xs text-base-content/70">
              <span v-if="deployStatus === 'installing'">Performing code pull, compose template rendering, and container
                builds...</span>
              <span v-else-if="deployStatus === 'success'">Container deployed successfully! Webserver proxy has been
                registered.</span>
              <span v-else>Deployment failed.See error logs below: </span>
            </p>

            <!-- Steps logs console -->
            <div class="space-y-4 mt-2 max-h-80 overflow-y-auto pr-1">
              <div v-for="(s, idx) in stepsLogs" :key="idx" class="bg-base-200 border border-base-300 rounded-lg p-3">
                <div class="flex items-center justify-between border-b border-base-300 pb-1.5 mb-2">
                  <span class="text-xs font-bold text-base-content flex items-center gap-2">
                    <span class="badge badge-sm badge-outline">{{ idx + 1 }}</span>
                    {{ s.name }}
                  </span>
                  <Icon :name="s.error ? 'mdi:close-circle' : 'mdi:check-circle'"
                    :class="s.error ? 'text-error' : 'text-success'" size="18" />
                </div>

                <div v-if="s.output" class="text-left">
                  <div class="text-[10px] font-semibold text-base-content/50 mb-1">Stdout:</div>
                  <pre
                    class="bg-base-300 text-[10px] font-mono p-2 rounded max-h-32 overflow-y-auto whitespace-pre-wrap">{{ s.output }}</pre>
                </div>

                <div v-if="s.error" class="text-left mt-2">
                  <div class="text-[10px] font-semibold text-error/60 mb-1">Stderr:</div>
                  <pre
                    class="bg-error/10 border border-error/20 text-error text-[10px] font-mono p-2 rounded whitespace-pre-wrap">{{ s.error }}</pre>
                </div>
              </div>
            </div>

            <div class="modal-action border-t border-base-300 pt-4 mt-6">
              <button @click="closeProgressModal" class="btn btn-sm" :disabled="deployStatus === 'installing'">
                Close Logs
              </button>
            </div>
          </div>
        </dialog>
      </template>
    </div>
  </div>
</template>

<script setup>
const route = useRoute();
const router = useRouter();
const slug = route.params.slug;

// State Variables
const loading = ref(true);
const error = ref('');
const step = ref(1);
const customMode = ref('image'); // image, compose, git

const requirements = ref({});
const deploying = ref(false);
const deployStatus = ref('installing'); // installing, success, error
const stepsLogs = ref([]);
const progressModal = ref(null);

const installForm = ref({
  path: '/opt/panel17/apps',
  domain: '',
  appName: '',
  vars: {},
  deploymentMethod: 'none',
  gitRepo: '',
  gitBranch: '',
  // custom UI
  image: '',
  compose_template: '',
  container_port: 80
});

onMounted(() => {
  loadManifestRequirements();
});

const loadManifestRequirements = async () => {
  try {
    loading.value = true;
    error.value = '';
    const res = await useNuxtApp().$axiosApi.get(`/containerapps/registry/${slug}/requirements`);
    if (res.data?.success) {
      requirements.value = res.data;

      // Auto pre-populate variables default values
      if (requirements.value.variables) {
        requirements.value.variables.forEach(v => {
          installForm.value.vars[v.key] = v.default || '';
        });
      }

      // Default domain name and app slug
      installForm.value.appName = slug === 'custom' ? 'my-container-app' : slug;
      installForm.value.domain = slug === 'custom' ? 'container.localhost' : `${slug}.localhost`;
    } else {
      throw new Error(res.data?.error || "Failed to fetch requirements");
    }
  } catch (err) {
    error.value = err.response?.data?.error || err.message || "Failed to parse registry configurations";
  } finally {
    loading.value = false;
  }
};

const goToSummary = () => {
  // Sync custom deployment methods
  if (slug === 'custom') {
    if (customMode.value === 'git') {
      installForm.value.deploymentMethod = 'git';
    } else {
      installForm.value.deploymentMethod = 'none';
    }
  }
  step.value = 2;
};

const launchDeployment = async () => {
  try {
    deploying.value = true;
    deployStatus.value = 'installing';
    stepsLogs.value = [];

    // Open console modal
    if (progressModal.value) {
      progressModal.value.showModal();
    }

    const payload = {
      ...installForm.value
    };

    const response = await useNuxtApp().$axiosApi.post(`/containerapps/registry/${slug}/install`, payload);

    if (response.data.success) {
      deployStatus.value = 'success';
      stepsLogs.value = response.data.steps || [{ name: "Podman compose deployment", output: "Container deployed successfully" }];
      useNuxtApp().$toast?.success("Container stack deployed successfully!");
    } else {
      deployStatus.value = 'error';
      stepsLogs.value = response.data.steps || [{ name: "Deployment run", error: response.data.error || "Execution error" }];
    }
  } catch (err) {
    deployStatus.value = 'error';
    const stepLogsErr = err.response?.data?.steps || [{ name: "API run", error: err.response?.data?.error || err.message || "Connection refused" }];
    stepsLogs.value = stepLogsErr;
  } finally {
    deploying.value = false;
  }
};

const closeProgressModal = () => {
  if (progressModal.value) {
    progressModal.value.close();
  }
  if (deployStatus.value === 'success') {
    // Redirect to list of applications
    router.push('/applications');
  }
};
</script>

<style scoped>
.textarea {
  resize: vertical;
}
</style>
