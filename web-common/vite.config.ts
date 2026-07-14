import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";
import type { Alias } from "vite";
import { svelteTesting } from "@testing-library/svelte/vite";

const alias: Alias[] = [
  {
    find: "src",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-common",
    replacement: "/../web-common/src",
  },
];

export default defineConfig(({ mode }) => {
  if (mode === "test") {
    alias.push({
      find: "$app/environment",
      replacement: "/../web-common/tests/app-environment.mock.ts",
    });
    // canvas-entity dynamically imports the admin client only in the cloud context; stub
    // it so web-common unit tests that pull in canvas-entity can resolve the import graph.
    alias.push({
      find: "@rilldata/web-admin/client",
      replacement: "/../web-common/tests/web-admin-client.mock.ts",
    });
  }

  return {
    resolve: {
      alias,
    },
    plugins: [sveltekit()],
    test: {
      projects: [
        {
          extends: "./vite.config.ts",
          plugins: [svelteTesting()],
          test: {
            name: "client",
            environment: "jsdom",
            clearMocks: true,
            setupFiles: ["./vitest-setup.ts"],
            globals: true,
            coverage: {
              provider: "v8",
              include: ["src/**"],
            },
          },
        },
      ],
    },
  };
});
