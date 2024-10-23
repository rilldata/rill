import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { viteSingleFile } from "vite-plugin-singlefile";
import dns from "dns";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

export default defineConfig({
  resolve: {
    alias: {
      "@rilldata/web-auth": "/src",
      "@rilldata/web-admin": "/../web-admin/src",
      "@rilldata/web-common": "/../web-common/src",
      "@rilldata/web-local": "/../web-local/src",
    },
  },
  plugins: [svelte(), viteSingleFile({ removeViteModuleLoader: true })],
  server: {
    port: 3000,
    strictPort: true,
  },

  build: {
    target: "es2019",
  },
});
