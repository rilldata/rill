import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import { defineConfig } from "vitest/config";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

const config = defineConfig({
  build: {
    rollupOptions: {
      // This ensures that the web-admin package is not bundled into the web-local package.
      // This is necessary because the Scheduled Reports dialog lives in `web-common` and imports the admin-client.
      external: (id) => id.startsWith("@rilldata/web-admin/"),
    },
  },
  resolve: {
    alias: {
      src: "/src", // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
      "@rilldata/web-admin": "/../web-admin/src",
    },
  },
  server: {
    strictPort: true,
    fs: {
      allow: ["."],
    },
  },
  define: {
    "import.meta.env.VITE_PLAYWRIGHT_TEST": process.env.PLAYWRIGHT_TEST,
  },
  plugins: [sveltekit()],
  envDir: "../",
  envPrefix: "RILL_UI_PUBLIC_",
});

export default config;
