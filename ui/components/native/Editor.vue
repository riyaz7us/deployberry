<template>
	<div>
  <quill-editor toolbar="full" contentType="html" :content="modelValue" @update:content="emit('update:modelValue',$event)" theme="snow"></quill-editor>
	</div>
</template>
<script setup>
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css';
const props = defineProps({
	modelValue: String
})
const emit = defineEmits(['update:modelValue']);
</script>
<style>
.ql-editor{max-height: 80vh;}
</style>