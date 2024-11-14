module.exports = {
  presets: [require("../web-common/tailwind.config.ts")],
  content: [
    "./src/**/*.{html,js,svelte,ts}",
    "../web-common/src/**/*.{html,js,svelte,ts}",
  ],
};
