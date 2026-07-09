/** @type {import('tailwindcss').Config} */
module.exports = {
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      "night",
      "coffee",
      "forest",
      "dracula",
      "light",
      "emerald",
      "lemonade",
      "winter"
    ],
  },
  content: ["./components/**/*.{js,vue,ts}", "./layouts/**/*.vue", "./pages/**/*.vue", "./plugins/**/*.{js,ts}", "./nuxt.config.{js,ts}", "presets/**/*.{js,vue,ts}"],
  theme: {
    extend: {},
    screens: {
      sm: "640px",
      // => @media (min-width: 640px) { ... }

      md: "768px",
      // => @media (min-width: 768px) { ... }

      lg: "1024px",
      // => @media (min-width: 1024px) { ... }

      xl: "1280px",
      // => @media (min-width: 1280px) { ... }

      "2xl": "1536px",
      // => @media (min-width: 1536px) { ... }
      xsAndDown: { raw: "screen and (max-width: 576px)" },
      smAndDown: { raw: "screen and (max-width: 768px)" },
      mdAndDown: { raw: "screen and (max-width: 992px)" },
    },
    height: {
      "10v": "10vh",
      "20v": "20vh",
      "30v": "30vh",
      "40v": "40vh",
      "50v": "50vh",
      "60v": "60vh",
      "70v": "70vh",
      "76v": "76vh",
      "80v": "80vh",
      "90v": "90vh",
      "100v": "100vh",
      screen: "100vh",
    },
  }
};
