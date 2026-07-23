<template>
  <div class="px-4 py-3 sm:px-6 sm:py-4 space-y-4">
    <h1 class="text-xl font-semibold text-base-content">Application Registry</h1>
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <nuxt-link :to="`/registry/${app.slug}`" class="card card-bordered bg-base-200 hover:bg-base-300 transition-all duration-200 flex flex-row items-start gap-4 p-4 cursor-pointer" v-for="app in apps" :key="app.slug">
        <Icon size="40" :name="app.icon||'mdi:application'" class="flex-shrink-0 text-primary" />
        <div class="min-w-0">
            <h2 class="font-bold text-base-content text-base truncate">{{ app.name }}</h2>
            <p class="text-xs text-base-content/60 mt-1 line-clamp-2">{{ app.description }}</p>
        </div>
      </nuxt-link>
    </div>
  </div>
</template>
<script setup>
useHead({ title: 'Application Registry' })
const apps = ref([]);
onMounted(() => {
  useNuxtApp()
    .$axiosApi.get("/registry/")
    .then((res) => {
      apps.value = res.data?.apps;
    });
});
</script>
