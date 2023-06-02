module.exports = {
	content: ['./src/**/*.{html,js,svelte,ts}'],
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
