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

const config = defineConfig({
  resolve: {
    alias: {
      src: "/src", // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
      "@shoelace-style/shoelace/dist/internal":
        "/../node_modules/@shoelace-style/shoelace/dist/internal",
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
