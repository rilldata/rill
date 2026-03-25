import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),

  // Suppress state_referenced_locally warnings from SvelteKit's auto-generated root.svelte.
  // Fixed upstream in SvelteKit 2.49.1 (https://github.com/sveltejs/kit/pull/15013).
  onwarn(warning, defaultHandler) {
    if (
      warning.code === "state_referenced_locally" &&
      warning.filename?.includes(".svelte-kit/generated/root.svelte")
    )
      return;
    defaultHandler(warning);
  },

  kit: {
    // CSP hash mode: SvelteKit computes SHA-256 hashes of inline scripts at
    // build time and injects them into a <meta http-equiv="Content-Security-Policy">
    // tag in each HTML page. This removes the need for 'unsafe-inline' in
    // script-src. frame-ancestors is not supported in <meta> and is therefore
    // kept in the Netlify HTTP headers (netlify.toml).
    csp: {
      mode: "hash",
      directives: {
        "default-src": ["self"],
        "script-src": [
          "self",
          "unsafe-eval",
          "https://*.app-us1.com/",
          "https://*.usepylon.com",
          "https://*.pusher.com",
        ],
        // style-src keeps 'unsafe-inline': runtime style injection from
        // CodeMirror and other libraries cannot be hash-attributed.
        "style-src": ["self", "unsafe-inline", "https://*.usepylon.com"],
        "img-src": ["https:", "data:", "blob:"],
        "frame-src": [
          "self",
          "https://www.youtube.com/",
          "https://www.loom.com/",
          "https://www.vimeo.com",
          "https://portal.withorb.com",
          "blob:",
          "data:",
        ],
        "form-action": ["self"],
        "object-src": ["none"],
        "base-uri": ["self"],
        "connect-src": [
          "self",
          "https://*.rilldata.com",
          "https://*.rilldata.io",
          "https://*.rilldata.in",
          "https://*.usepylon.com",
          "https://docs.google.com",
          "https://storage.googleapis.com",
          "https://cdn.prod.website-files.com",
          "https://*.stripe.com",
          "wss://*.pusher.com",
        ],
        "font-src": [
          "self",
          "https://fonts.gstatic.com",
          "https://*.usepylon.com",
        ],
      },
    },
    adapter: adapter({
      fallback: "index.html",
    }),
    files: {
      assets: "../web-common/static",
    },
  },
};

export default config;
