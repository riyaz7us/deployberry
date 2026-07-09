<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6">
    <div class="flex items-center justify-between border-b border-base-300 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">Container Registry</h1>
        <p class="text-sm text-base-content/60 mt-1">Deploy pre-packaged manifests or launch custom Docker/Podman containers.</p>
      </div>
      <nuxt-link to="/containerapps/custom" class="btn btn-primary btn-sm sm:btn-md gap-2">
        <Icon name="mdi:plus-circle" size="20" />
        Deploy Custom App
      </nuxt-link>
    </div>

    <!-- Error State -->
    <div v-if="error" class="alert alert-error">
      <Icon name="mdi:alert-circle" class="w-5 h-5 shrink-0" />
      <span>{{ error }}</span>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-64">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>

    <!-- Apps Grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <!-- Deploy Custom App card -->
      <nuxt-link 
        to="/containerapps/custom" 
        class="card card-bordered bg-base-100 hover:shadow-lg transition-all duration-200 border-dashed border-2 border-primary/40 hover:border-primary flex flex-col justify-between p-5 h-48 group cursor-pointer"
      >
        <div class="flex items-start gap-4">
          <div class="bg-primary/10 p-3 rounded-xl text-primary group-hover:bg-primary group-hover:text-primary-content transition-colors">
            <Icon size="36" name="mdi:docker" />
          </div>
          <div class="min-w-0">
            <h2 class="font-bold text-base-content text-base truncate group-hover:text-primary transition-colors">Custom App</h2>
            <p class="text-xs text-base-content/60 mt-1 line-clamp-3">Run any pre-built image, write a raw Dockerfile, or deploy a custom Docker Compose template.</p>
          </div>
        </div>
        <div class="text-xs text-primary font-medium flex items-center gap-1 mt-auto">
          Deploy Custom App <Icon name="mdi:arrow-right" class="w-4 h-4 transition-transform group-hover:translate-x-1" />
        </div>
      </nuxt-link>

      <!-- Pre-packaged App cards -->
      <nuxt-link 
        :to="`/containerapps/${app.slug}`" 
        class="card card-bordered bg-base-100 hover:shadow-lg hover:border-primary transition-all duration-200 flex flex-col justify-between p-5 h-48 group cursor-pointer" 
        v-for="app in apps" 
        :key="app.slug"
      >
        <div class="flex items-start gap-4">
          <div class="bg-base-200 p-3 rounded-xl text-primary group-hover:bg-primary group-hover:text-primary-content transition-colors">
            <Icon size="36" :name="app.icon || 'mdi:package-variant'" />
          </div>
          <div class="min-w-0">
            <h2 class="font-bold text-base-content text-base truncate group-hover:text-primary transition-colors">{{ app.name }}</h2>
            <p class="text-xs text-base-content/60 mt-1 line-clamp-3">{{ app.description }}</p>
          </div>
        </div>
        
        <div class="flex flex-wrap gap-1 mt-3">
          <span 
            v-for="tag in app.tags || []" 
            :key="tag" 
            class="badge badge-sm badge-outline text-[10px] text-base-content/75"
          >
            {{ tag }}
          </span>
        </div>
      </nuxt-link>
    </div>
  </div>
</template>

<script setup>
const apps = ref([]);
const loading = ref(true);
const error = ref('');

onMounted(async () => {
  try {
    loading.value = true;
    error.value = '';
    const res = await useNuxtApp().$axiosApi.get("/containerapps/registry");
    if (res.data?.success) {
      apps.value = res.data?.apps || [];
    } else {
      throw new Error(res.data?.error || "Failed to fetch registry applications");
    }
  } catch (err) {
    error.value = err.response?.data?.error || err.message || "Failed to load container registry";
  } finally {
    loading.value = false;
  }
});
</script>
