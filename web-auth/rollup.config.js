import { spawn } from "child_process";
import fs from "fs";
import path from "path";
import svelte from "rollup-plugin-svelte";
import commonjs from "@rollup/plugin-commonjs";
import resolve from "@rollup/plugin-node-resolve";
import postcss from "rollup-plugin-postcss";
import alias from "@rollup/plugin-alias";
import sveltePreprocess from "svelte-preprocess";
import postcssImport from "postcss-import";
import autoprefixer from "autoprefixer";
import tailwind from "tailwindcss";
import purgeCss from "@fullhuman/postcss-purgecss";
import dotenv from "dotenv";

dotenv.config();
const production = !process.env.ROLLUP_WATCH;

const environmentVars = [
  "VITE_RILL_CLOUD_AUTH0_CLIENT_IDS",
  "VITE_DISABLE_FORGOT_PASS_DOMAINS",
  "VITE_CONNECTION_MAP",
];

const removeUnusedCss = purgeCss({
  content: [
    "./src/**/*.html",
    "./src/**/*.svelte",
    "../web-common/**/*.{html,js,svelte,ts}",
  ],
  defaultExtractor: (content) => content.match(/[A-Za-z0-9-_:/]+/g) || [],
});

function inlineSvelte(template) {
  return {
    name: "Svelte Inliner",
    generateBundle(opts, bundle) {
      const file = path.parse(opts.file).base;
      const code = bundle[file].code;
      const output = fs.readFileSync(template, "utf-8");

      // Replace script tag with svelte component bundle
      bundle[file].code = output.replace("%%script%%", () => code);

      // Replace all environment variables
      environmentVars.forEach((envVar) => {
        bundle[file].code = bundle[file].code.replace(`"%%${envVar}%%"`, () =>
          JSON.stringify(process.env[envVar])
        );
      });
    },
  };
}

function serve() {
  let server;

  function toExit() {
    if (server) server.kill(0);
  }

  return {
    writeBundle() {
      if (server) return;
      server = spawn("npm", ["run", "start", "--", "--dev"], {
        stdio: ["ignore", "inherit", "inherit"],
        shell: true,
      });

      process.on("SIGTERM", toExit);
      process.on("exit", toExit);
    },
  };
}

export default {
  input: "src/exporter.ts",
  output: {
    sourcemap: false,
    format: "iife",
    name: "app",
    file: "bundle.html",
  },
  plugins: [
    alias({
      entries: [
        { find: "@rilldata/web-common", replacement: "../web-common/src" },
        { find: "@rilldata/web-admin", replacement: "../web-admin/src" },
        { find: "@rilldata/web-local", replacement: "../web-local/src" },
      ],
    }),
    svelte({
      // This tells svelte to run some preprocessing
      preprocess: sveltePreprocess({
        postcss: true, // And tells it to specifically run postcss!
      }),
      compilerOptions: {
        // enable run-time checks when not in production
        dev: !production,
      },
    }),
    postcss({
      plugins: [
        postcssImport,
        tailwind(),
        autoprefixer,
        production && removeUnusedCss,
      ].filter(Boolean),
    }),
    resolve({
      browser: true,
      dedupe: ["svelte"],
      exportConditions: ["svelte"],
      extensions: [".js", ".svelte", ".ts"],
    }),
    commonjs(),

    // In dev mode, call `npm run start` once
    // the bundle has been generated
    !production && serve(),
    inlineSvelte("./src/template.html"),
  ],
  watch: {
    clearScreen: false,
  },
};
