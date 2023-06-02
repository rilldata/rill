module.exports = {
  // need to add this for storybook
  // https://www.kantega.no/blogg/setting-up-storybook-7-with-vite-and-tailwind-css
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,svelte}",
  ],
  /** Once we have applied dark styling to all UI elements, remove this line */
  darkMode: "class",
  theme: {
    extend: {},
  },
  plugins: [],
  safelist: [
    "text-blue-800", // needed for blue text in filter pills
    "ui-copy-code", // needed for code in measure expressions
  ],
};
