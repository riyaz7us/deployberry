<template>
  <div class="shadow w-fit p-2">
    <img v-if="image" style="max-width:200px" :src="`//oscrm.test/storage/${image}`" /><br />
    <button @click="$refs['imageUpload'].click()" class="btn btn-primary btn-ghost p-1 w-full">Upload</button>
    <input type="file" hidden ref="imageUpload" id="Image" label="image" @input="uploadImage($event)" />
  </div>
</template>

<script setup>
const props = defineProps({
  image: { required: true, default: null },
  folder: { required: false, default: "images" },
});
const emit = defineEmits(["uploaded"]);
function uploadImage(e) {
  let image = e.target.files[0];
  e.target.value = "";
  if (image === null) {
    return;
  }
  let formdata = new FormData();
  formdata.append("image", image);
  formdata.append("path", this.folder);
  //formdata.append("delete", this.school_logo);
  useNuxtApp().$axiosApi
    .post(`auth/images`, formdata, {
      headers: { "Content-type": "multipart/form-data" },
    })
    .then(
      (res) => {
        this.cover_image = res.data.image;
        emit("uploaded", res.data.image);
      },
      (err) => {}
    );
}
</script>

<style></style>
