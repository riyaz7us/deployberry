import { useAuthStore } from "~/store/index";

export default defineNuxtPlugin({
  parallel: true,
  setup(nuxtApp) {
    const defaultUrl = nuxtApp.$config.public.BASE_URL;

    const baseFetch = $fetch.create({
      baseURL: defaultUrl,
      onRequest({ options }) {
        let token = localStorage.getItem("token");
        if (token) {
          options.headers = new Headers(options.headers || {});
          options.headers.set("Authorization", "Bearer " + token);
        }
      },
      onResponseError({ response }) {
        console.log(response);
        if (response.status === 401) {
          navigateTo("/login");
        }
        else if (response.status >= 499) {
          useNuxtApp().$bus.$emit("toaster", ["An Internal Error Occured!", "danger"]);
        } else {
          let errs;
          let rData = response._data || {};
          if (rData.error) {
            errs = rData.error;
          } else if (rData.message) {
            errs = rData.message;
          } else {
            errs = rData;
          }
          if (typeof errs === "object") {
            let errMsg = "";
            Object.keys(errs).forEach((ei) => {
              if (Array.isArray(errs[ei])) {
                errMsg += errs[ei].join("<br/>") + "<br>";
              } else {
                errMsg += errs[ei] + "<br>";
              }
            });
            useNuxtApp().$bus.$emit("toaster", [errMsg, "danger"]);
          } else {
            useNuxtApp().$bus.$emit("toaster", [errs, "danger"]);
          }
        }
      }
    });

    const axiosApi = {
      get: (url, config) => baseFetch(url, { method: 'GET', ...config }).then(data => ({ data })),
      post: (url, body, config) => baseFetch(url, { method: 'POST', body, ...config }).then(data => ({ data })),
      put: (url, body, config) => baseFetch(url, { method: 'PUT', body, ...config }).then(data => ({ data })),
      delete: (url, config) => baseFetch(url, { method: 'DELETE', ...config }).then(data => ({ data }))
    };

    return {
      provide: {
        axiosApi: axiosApi,
      },
    };
  },
});
