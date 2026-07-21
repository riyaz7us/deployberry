import { useAuthStore } from "~/store/index";

export default defineNuxtPlugin({
  parallel: true,
  setup(nuxtApp) {
    let defaultUrl = nuxtApp.$config.public.BASE_URL;

    if (typeof window !== "undefined" && !import.meta.dev) {
      defaultUrl = window.location.origin + "/api";
    }

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

    const sseConnect = (url, onMessage, onError) => {
      let token = localStorage.getItem("token");
      let fullUrl = defaultUrl + (url.startsWith('/') ? url : '/' + url);

      const abortController = new AbortController();

      const startStream = async () => {
        try {
          const response = await fetch(fullUrl, {
            headers: {
              'Authorization': token ? `Bearer ${token}` : '',
              'Accept': 'text/event-stream',
            },
            signal: abortController.signal
          });

          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }

          const reader = response.body.getReader();
          const decoder = new TextDecoder();
          let buffer = '';

          while (true) {
            const { value, done } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');

            // Keep the last line in the buffer in case it is incomplete
            buffer = lines.pop();

            for (const line of lines) {
              const trimmed = line.trim();
              if (trimmed.startsWith('data:')) {
                const dataStr = trimmed.slice(5).trim();
                if (onMessage) {
                  onMessage(dataStr);
                }
              }
            }
          }
        } catch (error) {
          if (error.name !== 'AbortError') {
            if (onError) onError(error);
          }
        }
      };

      startStream();

      return {
        close: () => {
          abortController.abort();
        }
      };
    };

    return {
      provide: {
        axiosApi: axiosApi,
        sseConnect: sseConnect,
      },
    };
  },
});
