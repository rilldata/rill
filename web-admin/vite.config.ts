import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import { defineConfig } from "vitest/config";
import { readPublicEmailDomains } from "./src/features/projects/user-management/readPublicEmailDomains";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

export default defineConfig({
  resolve: {
    alias: {
      "@rilldata/web-admin": "/src",
      "@rilldata/web-common": "/../web-common/src",
    },
  },
  server: {
    port: 3000,
    strictPort: true,
  },
  define: {
    RillPublicEmailDomains: readPublicEmailDomains(),
  },
  plugins: [sveltekit()],
  envDir: "../",
  envPrefix: "RILL_UI_PUBLIC_",
});
