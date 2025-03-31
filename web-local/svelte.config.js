import adapter from "@sveltejs/adapter-static";
import preprocess from "svelte-preprocess";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: preprocess(),

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
