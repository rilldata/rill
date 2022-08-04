import adapter from "@sveltejs/adapter-static";
import { execSync } from "child_process";
import { readFileSync } from "fs";
import { resolve } from "path";
import preprocess from "svelte-preprocess";
import { fileURLToPath } from "url";

// get current version
const file = fileURLToPath(new URL("package.json", import.meta.url));
const json = readFileSync(file, "utf8");
const pkg = JSON.parse(json);

// attempt to get current commit hash
let commitHash = "";
try {
  commitHash = execSync("git rev-parse --short HEAD").toString().trim()
} catch (e) {
  console.log("Could not get commit hash - most likely not in a git repo");
}

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
