/** @type {import('tailwindcss').Config} */

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
    extend: {
      colors: {
        blue: {
          50: "var(--color-primary-50)",
          100: "var(--color-primary-100)",
          200: "var(--color-primary-200)",
          300: "var(--color-primary-300)",
          400: "var(--color-primary-400)",
          500: "var(--color-primary-500)",
          600: "var(--color-primary-600)",
          700: "var(--color-primary-700)",
          800: "var(--color-primary-800)",
          900: "var(--color-primary-900)",
          950: "var(--color-primary-950)",
        },
      }
    },
  },
  plugins: [],
  safelist: [
    "text-blue-800", // needed for blue text in filter pills
    "ui-copy-code", // needed for code in measure expressions
  ],
};
