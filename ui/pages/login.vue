<template>
  <div class="min-h-screen bg-base-100 p-4 grid items-center justify-center">
    <div class="w-full max-w-md">
      <!-- Login Card -->
      <div class="card bg-base-200 border border-base-300 shadow-xl">
        <div class="card-body p-8">
          <!-- Header -->
          <div class="text-center mb-8">
              <img @click="navigateTo('/')" class="mx-auto cursor-pointer" src="/newlogo.svg" width="260" />
            <p class="text-base-content/60 mt-2">Sign in to your account</p>
          </div>

          <!-- Login Form -->
          <form @submit.prevent="login" class="space-y-6">
            <!-- Username Field -->
            <NativeTextField
              v-model="loginData.username"
              label="Username"
              placeholder="Enter your username"
              required
            />

            <!-- Password Field -->
            <NativeTextField
              v-model="loginData.password"
              type="password"
              label="Password"
              placeholder="Enter your password"
              required
            />

            <!-- Login Button -->
            <div class="form-control mt-8">
              <button
                type="submit"
                :disabled="load"
                class="btn btn-primary w-full"
                :class="{ 'loading': load }"
              >
                <span v-if="!load">Sign In</span>
                <span v-else>Please Wait...</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
definePageMeta({
  layout: "plain",
});
const { $login } = useNuxtApp();
const loginData = ref({ username: '', password: '' });
const load = ref(false);

async function login() {
  load.value = true;
  try {
    await $login(loginData.value.username, loginData.value.password);
    navigateTo('/');
    // Redirect after login if needed
  } catch (error) {
    load.value = false;
  }
}
</script>

<style scoped>
/* Custom styles for enhanced login experience */
.login-card {
  backdrop-filter: blur(10px);
}

/* Focus states for better accessibility */
.input:focus {
  box-shadow: 0 0 0 2px rgba(var(--primary), 0.2);
}
</style>
