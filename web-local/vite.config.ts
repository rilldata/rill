import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import { defineConfig } from "vite";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

let runtimeUrl = "";
try {
  runtimeUrl = process.env.RILL_DEV ? "http://localhost:9009" : "";
} catch (e) {
  console.error(e);
}

const config = defineConfig(({ mode }) => ({
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
      // Adding $img alias to fix Vite build warnings due to static assets referenced in CSS
      // See: https://stackoverflow.com/questions/75843825/sveltekit-dev-build-and-path-problems-with-static-assets-referenced-in-css
      $img: mode === "production" ? "/../web-common/static/img" : "../img",
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
    "import.meta.env.VITE_PLAYWRIGHT_TEST": process.env.PLAYWRIGHT_TEST,
  },
  plugins: [sveltekit()],
}));

export default config;
