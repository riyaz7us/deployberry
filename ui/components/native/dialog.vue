<template>
	<div v-show="modelValue"
		style="z-index: 999;position:fixed;top:0;left:0;width:100%;height:100%;display: flex;padding:20px;justify-content: center;"
		class="dialog">
		<div style="background:#00000055;backdrop-filter: blur(5px);" class="h-full w-full bg-red-500 absolute"
			@click="modelValue = false"></div>
		<div @click:outside="modelValue = false" style="margin:auto;padding:30px;border-radius: 5px;"
			class="relative max-w-5xl bg-base-100">
			<Icon @click="modelValue = false" name="mdi:close"
				class="absolute cursor-pointer top-1 right-1 bg-red-500 rounded-full text-black" />
			<div class="max-h-[80vh] overflow-y-auto p-2" style="scrollbar-width: thin;-webkit-scrollbar-width: thin;">
				<slot />
			</div>
		</div>
	</div>
</template>
<script setup>
const modelValue = defineModel();
const props = defineProps({
	background: {
		default: "var(--bg-base-300)"
	}
})
</script>