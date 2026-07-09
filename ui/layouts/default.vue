<template>
  <main class="min-h-screen bg-base-100 text-base-content">
    <header
      class="text-base-content flex justify-between items-center p-3 border-b border-base-200 bg-base-100/80 backdrop-blur-md sticky top-0 z-40">
      <div class="flex items-center gap-2 mr-auto py-1">
        <button class="btn btn-ghost btn-sm lg:hidden p-1 min-h-0 h-auto" @click="navdrawer = !navdrawer"
          aria-label="Toggle navigation menu">
          <Icon name="mdi:menu" class="text-2xl" />
        </button>
        <img @click="navigateTo('/')" class="cursor-pointer" src="/newlogo.svg" width="160" alt="Logo" />
      </div>
      <div class="flex gap-2 items-center">
        <button class="btn btn-outline btn-xs flex items-center gap-2" v-if="!home" @click="goBack()">
          <Icon name="mdi:arrow-left" /> Back
        </button>
        <Icon name="mdi:palette" @click="changeColor()"
          class="text-2xl cursor-pointer text-primary hover:scale-110 transition-transform" />
      </div>
    </header>
    <NuxtSnackbar />
    <div class="flex flex-col lg:grid lg:grid-cols-12 min-h-[calc(100vh-60px)]">
      <!-- Sidebar Desktop -->
      <div class="hidden lg:block lg:col-span-2 border-r border-base-200 bg-base-100/50 backdrop-blur-sm">
        <ul class="menu p-4 space-y-1 text-base-content">
          <li v-for="link in navLinks" :key="link.name">
            <NuxtLink :to="link.url" class="flex items-center gap-3 py-2 px-3 rounded-lg"
              active-class="active bg-primary text-primary-content font-medium">
              <Icon :name="link.icon" size="24" />
              <span>{{ link.name }}</span>
            </NuxtLink>
          </li>
        </ul>
      </div>

      <!-- Mobile Sidebar Drawer (Overlay) -->
      <Transition name="drawer-slide">
        <div v-if="navdrawer" class="lg:hidden fixed inset-0 z-50 flex">
          <!-- Overlay backdrop -->
          <div class="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity duration-300"
            @click="navdrawer = false"></div>
          <!-- Sidebar content drawer -->
          <div
            class="relative flex flex-col w-72 max-w-[80vw] h-full bg-base-100 p-4 shadow-2xl border-r border-base-200 transition-transform duration-300 ease-out">
            <div class="flex items-center justify-between mb-6 pb-4 border-b border-base-200">
              <img @click="navigateTo('/'); navdrawer = false" class="cursor-pointer" src="/newlogo.svg" width="130"
                alt="Logo" />
              <button class="btn btn-ghost btn-sm p-1" @click="navdrawer = false" aria-label="Close menu">
                <Icon name="mdi:close" class="text-2xl" />
              </button>
            </div>
            <ul class="menu space-y-1 text-base-content overflow-y-auto flex-1">
              <li v-for="link in navLinks" :key="link.name">
                <NuxtLink :to="link.url" @click="navdrawer = false"
                  class="flex items-center gap-3 py-3 px-4 rounded-lg hover:bg-base-200"
                  active-class="active bg-primary text-primary-content font-medium">
                  <Icon :name="link.icon" size="24" />
                  <span>{{ link.name }}</span>
                </NuxtLink>
              </li>
            </ul>
          </div>
        </div>
      </Transition>

      <!-- Main Content -->
      <div class="col-span-12 lg:col-span-10 min-h-0">
        <slot />
      </div>
    </div>
  </main>
</template>
<script setup>
const route = useRoute();
const home = ref(false);
const navLinks = [
  { name: "Dashboard", url: "/", icon: "mdi:view-dashboard" },
  { name: "File Manager", url: "/filemanager", icon: "fluent:folder-open-vertical-24-filled" },
  { name: "AppRunner", url: "/registry", icon: "mdi:package-variant" },
  { name: "Containers", url: "/containerapps", icon: "mdi:docker" },
  { name: "Applications", url: "/applications", icon: "mdi:package-variant" },
  { name: "Nginx Manager", url: "/nginxmanager", icon: "skill-icons:nginx" },
  { name: "Caddy Manager", url: "/caddymanager", icon: "mdi:web-plus" },
  { name: "SQL Manager", url: "/sqlmanager", icon: "skill-icons:mysql-dark" },
  { name: "Process Manager", url: "/pm2", icon: "skill-icons:nodejs-dark" },
  { name: "Languages", url: "/languages", icon: "mdi:code-block-braces" },
  { name: "Databases", url: "/databases", icon: "mdi:database" },
];
const navdrawer = ref(false);
const nextColor = ref(0);
watch(
  () => route.path,
  () => {
    if (route.path == "/") {
      home.value = true;
    } else {
      home.value = false;
    }
  }
);

const colors = ref(["night", "luxury", "sunset", "dracula", "silk", "lemonade", "cupcake"]);

function changeColor() {
  // Cycle to next theme
  nextColor.value = (nextColor.value + 1) % colors.value.length;
  const themeName = colors.value[nextColor.value];

  // Set DaisyUI theme
  document.documentElement.setAttribute("data-theme", themeName);

  // Save to localStorage
  if (process.client) {
    localStorage.setItem("daisyui-theme", themeName);
    localStorage.setItem("theme-index", nextColor.value.toString());
  }

  console.log("Switched to theme:", themeName, "at index:", nextColor.value);
}

onMounted(() => {
  if (route.path == "/") {
    home.value = true;
  }

  // Load saved theme or set default
  if (process.client) {
    const savedTheme = localStorage.getItem("daisyui-theme");
    const savedIndex = localStorage.getItem("theme-index");

    if (savedTheme && colors.value.includes(savedTheme)) {
      const themeIndex = colors.value.indexOf(savedTheme);
      nextColor.value = themeIndex;
      document.documentElement.setAttribute("data-theme", savedTheme);
      console.log("Loaded saved theme:", savedTheme);
    } else {
      // Set default theme (first one)
      const defaultTheme = colors.value[0];
      document.documentElement.setAttribute("data-theme", defaultTheme);
      console.log("Set default theme:", defaultTheme);
    }
  }
});
async function addUser() {
  await this.$api.post("/register", regData).then(
    (res) => {
      //console.log("🚀 addUser ~ res:", res);
    },
    (err) => {
      //console.log("🚀 addUser ~ err:", err);
    }
  );
  // Handle response...
}
async function checkLogin() {
  if (useAuthStore().$state.user) {
    navigateTo("/account");
  } else {
    useNuxtApp().$bus.$emit("loginDialog", true);
  }
}

function goBack() {
  useRouter().back();
}
</script>
<style lang="scss">
/* Drawer slide transitions */
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: opacity 0.3s ease;

  .relative {
    transition: transform 0.3s ease-out;
  }
}

.drawer-slide-enter-from,
.drawer-slide-leave-to {
  opacity: 0;

  .relative {
    transform: translateX(-100%);
  }
}

.register .m-dialog-layout {
  background: #000000cc;
  backdrop-filter: blur(5px);
}

.navbarBg {
  background: #006b52aa !important;
  /*background: linear-gradient(120deg, #ebb30bdd 10%, #e5b804aa) !important;*/
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
}

.menuBtn {
  padding: 2px 5px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  backdrop-filter: blur(2px);
  font-weight: bold;
  color: white;
}

.menuBtn:hover {
  color: white !important;
  background: var(--primary-900);
}

.navbar {
  background: #ffffff00;
  width: 100%;
  padding: 12px;
  z-index: 777;
}
</style>
