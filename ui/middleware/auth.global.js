export default defineNuxtRouteMiddleware((to, from) => {
  return useNuxtApp()
  .$checkLogin()
  .then(
    (res) => {
      return;
    },
    (err) => {
          if(to.path != '/login'){
            return navigateTo('/login');
          }
        }
      );
});
