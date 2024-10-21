import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import { defineConfig } from "vitest/config";

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
  server: {
    port: 3000,
    strictPort: true,
  },
  plugins: [sveltekit()],
});
