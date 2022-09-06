import { sveltekit } from "@sveltejs/kit/vite";
import { execSync } from "child_process";
import dns from "dns";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { defineConfig } from "vite";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

// get current version
const file = fileURLToPath(new URL("package.json", import.meta.url));
const json = readFileSync(file, "utf8");
const pkg = JSON.parse(json);

// attempt to get current commit hash
let commitHash = "";
try {
  commitHash = execSync("git rev-parse --short HEAD").toString().trim();
} catch (e) {
  console.log("Could not get commit hash - most likely not in a git repo");
}

const config = defineConfig({
  resolve: {
    alias: {
      $common: "/src/common",
      $server: "/src/server",
    },
  },
  server: {
    port: 3000,
    strictPort: true,
  },
  define: {
    RILL_VERSION: `"${pkg.version}"`,
    RILL_COMMIT: `"${commitHash}"`,
  },
  plugins: [sveltekit()],
});

export default config;
