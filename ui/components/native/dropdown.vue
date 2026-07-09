<template>
  <div @keydown.esc="show = false" @focusout="closeDD($event)">
    <div class="relative">
      <!-- Dropdown toggle button -->
      <button @click="show = !show" class="flex items-center py-1 px-2 rounded-md border-2 btn btn-outline">
        {{ label }}
      </button>

      <!-- Dropdown menu -->
      <div @click="show = false" v-show="show" class="absolute left-0 shadow-2xl w-44 text-white flex flex-col backdrop-blur-lg p-1 z-50 text-left overflow-y-auto">
        <nuxt-link v-for="(s, si) in props.items" :key="si" @click="modelValue = s.link||s" class="block hover:bg-base-300 cursor-pointer p-2 my-1 text-sm"> {{ s.label||s }} </nuxt-link>
      </div>
    </div>
  </div>
</template>
<script setup>
const props = defineProps({
  items: { required: true },
  label:{required:true}
});
const modelValue = defineModel();
const show = ref(false);
function closeDD(e) {
  setTimeout(() => {
    show.value = false;
  }, 100);
}
</script>
<style scoped></style>
