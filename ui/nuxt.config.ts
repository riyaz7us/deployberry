// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
export default defineNuxtConfig({
  ssr: false,
  devServer: {
    port: 6601,
  },
  experimental: {
    emitRouteChunkError: "automatic",
  },

  app: {
    head: {
      htmlAttrs: {
        lang: "en",
      },
      charset: "utf-8",
      viewport: "width=device-width, initial-scale=1",
      title: "AppServer Control Panel",
      titleTemplate: "%s - Deployberry",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
        {
          name: "description",
          content: "Deployberry.",
        },
        { name: "og:title", content: "Deployberry" },
        { name: "twitter:title", content: "Deployberry" },
        { name: "twitter:card", content: "summary_large_image" },
        {
          name: "og:title",
          content: "Deployberry",
        },
        {
          name: "twitter:title",
          content: "Deployberry",
        },
        { name: "og:image", content: "https://Deployberry.spotverge.com/logo.png" },
        { name: "twitter:image", content: "https://Deployberry.spotverge.com/logo.png" },
      ],
      link: [
        { rel: "icon", type: "image/png", href: "/icon.png" },
        { rel: "preconnect", href: "https://fonts.googleapis.com" },
        { rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: "" },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap",
        },
      ],
      script: [
        //{ hid: "stripe", src: "https://js.stripe.com/v3/", defer: true, async: true },
        //{ hid: "umami", src: "https://analytics.eu.umami.is/script.js", "data-website-id": "", defer: true, async: true },
      ],
    },
  },

  css: ["@/styles.scss", "@/styles.css"],
  icon: {
    mode: "css",
    cssLayer: "base",
  },
  runtimeConfig: {
    public: {
      BASE_URL: process.env.BASE_URL,
      MEDIA_URL: process.env.MEDIA_URL,
      BACKEND_URL: process.env.BACKEND_URL,
    },
  },

  imports: {
    dirs: ["store"],
  },

  modules: [
    "@nuxtjs/google-fonts",
    "nuxt-snackbar",
    "@pinia/nuxt", //"@vite-pwa/nuxt",
    "@nuxt/icon",
  ],

  compatibilityDate: "2024-11-04",
  vite: {
    plugins: [tailwindcss()],
  },
  nitro: {
    output: {
      publicDir: "./dist",
    },
  },
});
