import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig, UserConfig } from "vitest/config";

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
  alias["$app/environment"] =
    "/../web-common/.storybook/app-environment.mock.ts";
}

export default defineConfig(({ mode }) => {
  if (mode === "test") {
    alias.push({
      find: /^svelte$/,
      replacement: "/../node_modules/svelte/index.mjs",
    });
  }

  return {
    resolve: {
      alias,
    },
    plugins: [svelte()],
    test: {
      coverage: {
        provider: "c8",
        src: ["./src"],
        all: true,
      },
      environment: "jsdom",
    },
  };
});
