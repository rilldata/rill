import { paraglideVitePlugin } from "@inlang/paraglide-js";
import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import { defineConfig } from "vitest/config";
import { readPublicEmailDomains } from "./src/features/projects/user-management/readPublicEmailDomains";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

export default defineConfig({
  test: {
    include: ["src/**/*.{test,spec}.{js,ts}"],
    coverage: {
      provider: "v8",
      reportsDirectory: "coverage",
      reporter: ["text", "lcov"],
      include: ["src/**/*.{js,ts,svelte}"],
      exclude: [
        "src/**/*.d.ts",
        "src/proto/gen/**",
        "src/client/gen/**",
        "src/runtime-client/**/gen/**",
        "src/lib/i18n/gen/**",
      ],
    },
  },
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
  preview: {
    port: 3000,
    strictPort: true,
  },
  define: {
    RillPublicEmailDomains: readPublicEmailDomains(),
  },
  optimizeDeps: {
    include: [
      "@tanstack/svelte-query",
      "@codemirror/view",
      "@codemirror/state",
      "@codemirror/language",
      "d3-scale",
      "d3-format",
      "d3-array",
      "luxon",
      "vega-lite",
      "memoize-weak",
    ],
    exclude: ["sveltekit-superforms"],
  },
  plugins: [
    sveltekit(),
    paraglideVitePlugin({
      project: "../web-common/src/lib/i18n/project.inlang",
      outdir: "../web-common/src/lib/i18n/gen",
      strategy: ["localStorage", "preferredLanguage", "baseLocale"],
    }),
  ],
  envDir: "../",
  envPrefix: "RILL_UI_PUBLIC_",
});
