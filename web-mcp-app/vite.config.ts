import { svelte } from "@sveltejs/vite-plugin-svelte";
import { fileURLToPath } from "node:url";
import { viteSingleFile } from "vite-plugin-singlefile";
import { defineConfig } from "vitest/config";

const here = (path: string) => fileURLToPath(new URL(path, import.meta.url));

// Modules that must never be kept for their side effects alone. Type-only imports of the
// runtime client are erased by TypeScript, but any stray value import of the barrel would
// otherwise drag the proto, Connect RPC and TanStack layers into the widget. These rules are
// a backstop; the real fix is keeping those imports out of the chart stack.
const sideEffectFreeModules = [
  /\/proto\/gen\//,
  /\/runtime-client\/v2\/gen\//,
  /node_modules\/@bufbuild\//,
  /node_modules\/@connectrpc\//,
  /node_modules\/@tanstack\//,
];

export default defineConfig({
  plugins: [svelte(), viteSingleFile({ removeViteModuleLoader: true })],
  resolve: {
    alias: {
      "@rilldata/web-common": here("../web-common/src"),
      "$app/environment": here("./src/stubs/app-environment.ts"),
    },
  },
  build: {
    target: "es2022",
    outDir: "dist",
    emptyOutDir: true,
    modulePreload: { polyfill: false },
    cssCodeSplit: false,
    rolldownOptions: {
      input: here("./widget.html"),
      treeshake: {
        moduleSideEffects: sideEffectFreeModules.map((test) => ({
          test,
          sideEffects: false,
        })),
      },
    },
  },
  server: {
    port: 5178,
    strictPort: false,
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./vitest-setup.ts"],
    include: ["src/**/*.spec.ts"],
  },
});
