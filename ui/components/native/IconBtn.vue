<template>
	<MazBtn :loading="loading" color="primary" :size="size" :outline="outline" @click="($event)=>{emit('click',$event)}"> <Icon :class="t?'':'mr-2'" :name="icon"/> 
		<slot></slot>
	</MazBtn>
</template>
<script setup>
import MazBtn from "maz-ui/components/MazBtn";
const emit = defineEmits(['click']);
const props = defineProps({
	loading:Boolean,
	icon:String,
	color:String,
	size:String,
	outline:Boolean,
	t: Boolean
})
</script>