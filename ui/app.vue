<template>
  <div>
    <NuxtLayout>
      <NuxtPage />
      <!--NuxtPwaManifest /-->
    </NuxtLayout>
  </div>
</template>
<script setup>
const snack = reactive({
  open: false,
  msg: "",
  color: "",
  icon: "",
});
const loginDialog = ref(false);
const aiDialog = ref(false);
const query = ref("");
onBeforeMount(() => {});
onMounted(() => {
  window.chatbot_id = 1;
  if (typeof document !== "undefined") {
    let ele = document.createElement("script");
    let ele2 = document.createElement("script");
    ele.setAttribute("src", "https://unpkg.com/ionicons@5.5.2/dist/ionicons/ionicons.esm.js");
    ele.setAttribute("type", "module");
    ele2.setAttribute("nomodule", "true");
    ele2.setAttribute("src", "https://unpkg.com/ionicons@5.5.2/dist/ionicons/ionicons.js");
    document.head.appendChild(ele);
    document.head.appendChild(ele2);
  }
 // useNuxtApp().$checkLogin();
  useNuxtApp().$bus.$on("toaster", (toast) => {
    serveToast(toast[0], toast[1],toast[2]);
  });
  useNuxtApp().$bus.$on("loginDialog", (v) => {
    loginDialog.value = v;
  });
  useNuxtApp().$bus.$on("aiDialog", (v) => {
    aiDialog.value = true;
    query.value = v;
  });
});
function serveToast(m, t, act) {
  //console.log("snack", m, t);
  snack.open = true;
  snack.msg = m;
  snack.action = act;
  //console.log("snack",snack);
  if (t == "danger") {
    snack.color = "red";
    snack.icon = "mdi:close-circle";
  } else if (t == "success") {
    snack.color = "green";
    snack.icon = "mdi:check-circle";
  } else {
    (snack.icon = "mdi:information"), (snack.color = undefined);
  }
}
</script>
