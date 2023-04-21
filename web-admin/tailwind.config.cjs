module.exports = {
  presets: [require("../web-common/tailwind.config.cjs")],
  content: [
    "./src/**/*.{html,js,svelte,ts}",
    "../web-common/**/*.{html,js,svelte,ts}",
  ],
};
