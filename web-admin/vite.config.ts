import { paraglideVitePlugin } from "@inlang/paraglide-js";
import { sveltekit } from "@sveltejs/kit/vite";
import { svelteTesting } from "@testing-library/svelte/vite";
import dns from "dns";
import { configDefaults, defineConfig } from "vitest/config";
import { readPublicEmailDomains } from "./src/features/projects/user-management/readPublicEmailDomains";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

export default defineConfig({
  test: {
    projects: [
      {
        extends: true,
        test: {
          name: "unit",
          exclude: [...configDefaults.exclude, "**/*.component.spec.ts"],
        },
      },
      {
        extends: true,
        plugins: [svelteTesting()],
        ssr: {
          noExternal: ["lucide-svelte", "bits-ui", "runed", "svelte-toolbelt"],
        },
        test: {
          name: "components",
          environment: "jsdom",
          include: ["src/**/*.component.spec.ts"],
        },
      },
    ],
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
