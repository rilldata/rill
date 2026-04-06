import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),

  // Suppress state_referenced_locally warnings from SvelteKit's auto-generated root.svelte.
  // Fixed upstream in SvelteKit 2.49.1 (https://github.com/sveltejs/kit/pull/15013).
  onwarn(warning, defaultHandler) {
    if (
      warning.code === "state_referenced_locally" &&
      warning.filename?.includes(".svelte-kit/generated/root.svelte")
    )
      return;
    defaultHandler(warning);
  },

  kit: {
    adapter: adapter({
      fallback: "index.html",
    }),
    files: {
      assets: "../web-common/static",
    },
  },
};

export default config;
