import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      src: "/src", // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
    },
  },
  plugins: [svelte()],
  test: {
    coverage: {
      provider: "c8",
      src: ["./src"],
      all: true
    }
  },
});
