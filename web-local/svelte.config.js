import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
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
    // TODO: enable CSP after addressing error in Pylon and Codemirror
    // csp: {
    //   directives: {
    //     "script-src": ["self"],
    //   },
    // },
    files: {
      assets: "../web-common/static",
    },
  },
};

export default config;
