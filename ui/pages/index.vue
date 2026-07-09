<template>
  <div class="p-4">
    <div class="grid grid-cols-12 gap-4">
      <section class="col-span-12 lg:col-span-9">
        <!-- Quick Apps -->

        <h1 class="text-xl mt-4 font-light">Quick Install</h1>
        <div class="grid gap-3 items-start grid-cols-1 sm:grid-cols-2 md:grid-cols-3">
          <nuxt-link v-for="(l, li) in favourites" :key="li" class="p-4 h-fit card card-bordered bg-base-200 hover:bg-base-300 transition-colors" :to="'/registry/' + l.slug">
            <span class="text-lg flex items-center"> <Icon size="40" :name="l.icon" />&emsp;{{ l.name }} </span>
          </nuxt-link>
        </div>
        <!-- Navigation Links -->
        <div class="space-y-4" v-for="(l, li) in links" :key="li">
          <h1 class="text-xl mt-4 font-light">{{ l.subheader }}</h1>
          <div class="grid gap-3 items-start grid-cols-1 sm:grid-cols-2">
            <nuxt-link v-for="(l, li) in l.children" :key="li" class="p-4 h-fit card card-bordered bg-base-200 hover:bg-base-300 transition-colors" :to="l.url">
              <span class="text-lg flex items-center"> <Icon size="40" :name="l.icon" />&emsp;{{ l.name }} </span>
            </nuxt-link>
          </div>
        </div>
      </section>
      <!-- Compact Stats Column -->
      <aside class="col-span-12 lg:col-span-3 card card-bordered bg-base-100 space-y-3" v-if="stats">
        <!-- CPU Usage -->
        <div class="p-3 card">
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center">
              <Icon name="mdi:cpu-64-bit" size="16" class="mr-1" />
              <span class="text-sm font-medium">CPU Average</span>
            </div>
            <span class="text-xs text-base-content/70">{{ averageCpu.toFixed(1) }}%</span>
          </div>
          <!-- CPU Core Grid -->
          <div class="grid grid-cols-4 gap-1">
            <div
              v-for="(usage, index) in stats.cpu"
              :key="index"
              class="min-h-6 rounded-xl flex items-center justify-center text-[10px] font-bold transition-all duration-300"
              :class="getCpuCoreClass(usage)"
              :title="`Core ${index + 1}: ${usage.toFixed(1)}%`"
            >
              {{ index + 1 }}
            </div>
          </div>
        </div>

        <!-- Memory Usage -->
        <div class="p-3 card">
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center">
              <div :style="{ color: getUsageBackground(stats.memory) }">
                <span class="text-6xl italic font-black">{{ stats.memory.toFixed(1) }}</span
                >&nbsp;<span class="text-xs text-base-content/70">% RAM used</span>
              </div>
            </div>
          </div>
          <!-- Memory Progress Bar -->
          <div class="w-full rounded-full min-h-1 bg-black overflow-hidden">
            <div class="h-full transition-all min-h-1 duration-500 rounded-full" :style="{ width: '100%', background: getUsageBackground(stats.memory) }"></div>
          </div>
        </div>

        <!-- Disk Usage -->
        <div class="p-3 card" v-if="stats.disk_usage !== undefined">
          <!-- Disk Progress Bar -->
          <div class="w-full bg-black text-white rounded-sm overflow-hidden">
            <div class="h-full px-2 flex items-center justify-between transition-all duration-500" :style="{ width: '100%', background: getUsageBackground(stats.disk_usage) }">
              <span class="text-sm font-medium flex items-center"><Icon name="mdi:harddisk" size="16" class="mr-1" /> Disk</span>
              <span class="text-xs">{{ stats.disk_usage.toFixed(1) }}%</span>
            </div>
          </div>
        </div>

        <!-- System Info -->
        <div class="p-4 card bg-gradient-to-br from-base-100 to-base-200 border border-base-300 shadow-lg hover:shadow-xl transition-all duration-300">
          <h3 class="text-sm font-medium mb-4 flex items-center bg-base-200/50 p-2 rounded-lg">
            <Icon name="mdi:information-outline" size="20" class="mr-2" />
            <span class="text-content">System Information</span>
          </h3>
          <div class="grid grid-cols-3 gap-3">
            <!-- RAM -->
            <div class="text-center p-1 rounded-lg bg-base-200/30 hover:bg-base-200/50 transition-all">
              <div class="text-4xl font-bold text-content-base">{{ (stats.total_memory / 1024).toFixed(1) }}</div>
              <div class="text-[10px] font-medium text-base-content/60">GB RAM</div>
            </div>

            <!-- CPU -->
            <div class="text-center p-1 rounded-lg bg-base-200/30 hover:bg-base-200/50 transition-all">
              <div class="text-4xl font-bold text-content-base">{{ stats.cpu.length }}</div>
              <div class="text-[10px] font-medium text-base-content/60">CPU CORES</div>
            </div>

            <!-- Storage -->
            <div class="text-center p-1 rounded-lg bg-base-200/30 hover:bg-base-200/50 transition-all">
              <div class="text-4xl font-bold text-content-base">{{ (stats.disk_total / 1024).toFixed(1) }}</div>
              <div class="text-[10px] font-medium text-base-content/60">GB STORAGE</div>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
const links = [
  {
    subheader: "Web Server Management",
    children: [
      { name: "Nginx Manager", url: "/nginxmanager", icon: "skill-icons:nginx" },
      { name: "Caddy Manager", url: "/caddymanager", icon: "mdi:web-plus" },
    ],
  },
  {
    subheader: "Apps, Files, and Processes",
    children: [
      { name: "Applications", url: "/applications", icon: "mdi:package-variant" },
      { name: "File Manager", url: "/filemanager", icon: "fluent:folder-open-vertical-24-filled" },
      { name: "Process Manager", url: "/pm2", icon: "skill-icons:nodejs-dark" },
    ],
  },
  {
    subheader: "Programming Tools",
    children: [
      { name: "Languages", url: "/languages", icon: "mdi:code-block-braces" },
      { name: "Databases", url: "/databases", icon: "mdi:database" },
    ],
  },
];

const favourites = [
  {
    slug: "ghost",
    name: "Ghost",
    icon: "jam:ghost-org",
    description: "Professional open-source publishing platform",
  },
  {
    slug: "umami",
    name: "Umami Analytics",
    icon: "simple-icons:umami",
    description: "Privacy-focused open-source web analytics",
  },
  {
    slug: "laravel",
    name: "Laravel",
    icon: "mdi:laravel",
    description: "PHP web application framework",
  },
  {
    slug: "wordpress",
    name: "WordPress",
    icon: "mdi:wordpress",
    description: "Open-source CMS powering 40% of the web",
  },
  {
    slug: "erpnext",
    name: "ERP Next",
    icon: "simple-icons:erpnext",
    description: "Powerful Open Source ERP Solution",
    tags: ["erp", "python"],
  },
];

const stats = ref(null);
const intv = ref(null);

const averageCpu = computed(() => {
  if (!stats.value?.cpu) return 0;
  return stats.value.cpu.reduce((a, b) => a + b, 0) / stats.value.cpu.length;
});

onMounted(() => {
  getStats();
  intv.value = setInterval(getStats, 5000);
});

function getStats() {
  useNuxtApp()
    .$axiosApi.get("/stats")
    .then(
      (res) => {
        stats.value = res.data;
      },
      (err) => {
        console.error("Error fetching stats:", err);
      },
    );
}

function getCpuCoreClass(usage) {
  if (usage > 80) return "bg-red-500 ";
  if (usage > 60) return "bg-orange-400";
  if (usage > 40) return "bg-yellow-400 text-black";
  if (usage < 20) return "bg-emerald-400 text-black";
  return "bg-purple-500";
}
function getUsageBackground(usage) {
  let color = "#00ff00";
  if (usage > 50) {
    color = "#Ffbf00";
  }
  if (usage > 80) {
    color = "#FFA500";
  }
  if (usage > 90) {
    color = "#f00";
  }

  return `linear-gradient(90deg, ${color} 0%, transparent ${Math.ceil(usage).toFixed()}%)`;
}

onBeforeUnmount(() => {
  clearInterval(intv.value);
});
</script>
