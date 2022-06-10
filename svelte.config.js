import adapter from "@sveltejs/adapter-static";
import preprocess from "svelte-preprocess";
import { resolve } from "path";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: preprocess(),

  kit: {
    adapter: adapter({
      fallback: "index.html",
    }),

    vite: {
      resolve: {
        alias: {
          $common: resolve("./src/common"),
          $lib: resolve("./src/lib"),
          $server: resolve("./src/server"),
        },
      },
    },
  },
};

export default config;
