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

let runtimeUrl = "";
try {
  runtimeUrl = process.env.RILL_DEV ? "http://localhost:9009" : "";
} catch (e) {
  console.error(e);
}

const config = defineConfig({
  resolve: {
    alias: {
      src: "/src", // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
    },
  },
  server: {
    port: 3000,
    strictPort: true,
    fs: {
      allow: ["."],
    },
  },
  define: {
    RILL_RUNTIME_URL: `"${runtimeUrl}"`,
  },
  plugins: [sveltekit()],
});

export default config;
