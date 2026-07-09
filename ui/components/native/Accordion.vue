<template>
	<div class=" col-span-full">
		<button class="btn btn-link" @click="modelValue=!modelValue">{{ name }} <Icon :name="modelValue ?'mdi:chevron-double-up':'mdi:chevron-double-down' "/> </button>
		<div v-show="modelValue" class="my-2">
			<slot/>
		</div>
	</div>
</template>
<script setup>
const props = defineProps({
	name:{default:null}
})
const modelValue = defineModel();
</script>
