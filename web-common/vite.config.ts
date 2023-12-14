import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig, UserConfig } from "vitest/config";
import Icons from "unplugin-icons/vite";

type Writeable<T> = { -readonly [P in keyof T]: T[P] };
type Alias = Writeable<UserConfig["resolve"]["alias"]>;

const alias: Alias = [
  {
    find: "src",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-local",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-common",
    replacement: "/../web-common/src",
  },
];

if (process.env["STORYBOOK_MODE"] === "true") {
  alias.push({
    find: "$app/environment",
    replacement: "/../web-common/.storybook/app-environment.mock.ts",
  });
}

export default defineConfig(({ mode }) => {
  if (mode === "test") {
    alias.push({
      find: "$app/environment",
      replacement: "/../web-common/.storybook/app-environment.mock.ts",
    });
  }

  return {
    resolve: {
      alias,
    },
    plugins: [
      svelte(),
      Icons({
        compiler: "svelte",
        autoInstall: true,
      }),
    ],
    test: {
      // This alias fixes `onMount` not getting called during vitest unit tests.
      // See: https://stackoverflow.com/questions/76577665/vitest-and-svelte-components-onmount
      alias: [{ find: /^svelte$/, replacement: "svelte/internal" }],
      coverage: {
        provider: "c8",
        src: ["./src"],
        all: true,
      },
      environment: "jsdom",
    },
  };
});
