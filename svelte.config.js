import adapter from "@sveltejs/adapter-static";
import { execSync } from "child_process";
import { readFileSync } from "fs";
import { resolve } from "path";
import preprocess from "svelte-preprocess";
import { fileURLToPath } from "url";

const commitHash = execSync("git rev-parse --short HEAD").toString().trim();

const file = fileURLToPath(new URL("package.json", import.meta.url));
const json = readFileSync(file, "utf8");
const pkg = JSON.parse(json);

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
      define: {
        RILL_VERSION: `"${pkg.version}"`,
        RILL_COMMIT: `"${commitHash}"`,
      },
    },
  },
};

export default config;
