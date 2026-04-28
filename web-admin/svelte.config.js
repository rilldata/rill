import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { config as dotenv } from "dotenv";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";

// svelte.config.js runs before Vite loads .env files, so we load manually.
// envDir in vite.config.ts points to the repo root ("../").
const __dirname = dirname(fileURLToPath(import.meta.url));
dotenv({ path: resolve(__dirname, "../.env"), override: false });

const dev = process.env.RILL_ADMIN_FRONTEND_URL?.includes("localhost");

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
          // Injected by the Pylon widget at runtime; exact subdomain is unknown without CSP reports.
          "https://*.app-us1.com/",
          // widget.usepylon.com is the confirmed entry point (initPylonWidget.ts).
          // The remaining *.usepylon.com and *.pusher.com entries below are loaded
          // dynamically by the Pylon widget — narrow further once CSP violation reports
          // confirm the exact subdomains.
          // https://support.usepylon.com/articles/5968160735-chat-widget-debugging-guide
          "https://widget.usepylon.com",
          "https://*.pusher.com",
          ...(dev ? ["http:"] : []),
          // Hash of the inline script injected by the Pylon chat widget at runtime.
          // If Pylon updates their widget, this hash may need to be refreshed.
          "sha256-q7DzCTpmdcQlqCarsIE22KTL5subp7TPBUdWqrL6HJw=",
        ],
        // style-src keeps 'unsafe-inline': runtime style injection from
        // CodeMirror and other libraries cannot be hash-attributed.
        "style-src": ["self", "unsafe-inline", "https://*.usepylon.com"],
        "img-src": [...(dev ? ["http:"] : []), "https:", "data:", "blob:"],
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
          "wss://*.pusher.com",
          ...(dev ? ["http://localhost:*", "ws://localhost:*"] : []),
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
