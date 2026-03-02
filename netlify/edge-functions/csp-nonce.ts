import type { Config, Context } from "@netlify/edge-functions";

/**
 * Generates a cryptographically random base64 nonce for use in CSP headers.
 */
function generateNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return btoa(String.fromCharCode(...bytes));
}

/**
 * Builds the full Content-Security-Policy header value.
 *
 * script-src uses 'strict-dynamic' with a per-request nonce so that:
 *  - The nonce anchors trust to SvelteKit's bootstrap scripts.
 *  - Scripts dynamically loaded by those trusted scripts are also trusted,
 *    making host allowlists unnecessary (and they are intentionally omitted).
 *  - 'unsafe-inline' is kept only as a fallback for browsers that do not
 *    understand 'strict-dynamic'; modern browsers ignore it when a nonce
 *    or 'strict-dynamic' is present.
 *  - 'unsafe-eval' is still required by CodeMirror.
 *
 * frame-ancestors differs between the default and embed/share routes.
 */
function buildCSP(nonce: string, path: string): string {
  const isEmbed = path.startsWith("/-/embed/");
  const isShare = /\/-\/share\//.test(path);
  const frameAncestors = isEmbed || isShare ? "https:" : "'self'";

  // Pylon CDN is kept in style-src (not script-src) for the default route
  // because Pylon injects styles. Embed/share routes don't load Pylon.
  const styleSrc =
    isEmbed || isShare
      ? "'self' 'unsafe-inline'"
      : "'self' 'unsafe-inline' https://*.usepylon.com";

  const directives = [
    "default-src 'self'",
    `script-src 'strict-dynamic' 'nonce-${nonce}' 'unsafe-inline' 'unsafe-eval'`,
    `style-src ${styleSrc}`,
    "img-src https: data: blob:",
    "frame-src 'self' https://www.youtube.com/ https://www.loom.com/ https://www.vimeo.com https://portal.withorb.com blob: data:",
    `frame-ancestors ${frameAncestors}`,
    "form-action 'self'",
    "object-src 'none'",
    "connect-src 'self' https://*.rilldata.com https://*.rilldata.io https://*.rilldata.in https://*.usepylon.com https://docs.google.com https://storage.googleapis.com https://cdn.prod.website-files.com https://*.stripe.com wss://*.pusher.com",
    "font-src 'self' https://fonts.gstatic.com https://*.usepylon.com",
  ];

  return directives.join("; ");
}

export default async function handler(
  request: Request,
  context: Context,
): Promise<Response> {
  const response = await context.next();

  // Only process HTML responses; pass everything else through unchanged.
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("text/html")) {
    return response;
  }

  const nonce = generateNonce();
  const { pathname } = new URL(request.url);
  const csp = buildCSP(nonce, pathname);

  // Inject the nonce attribute into every <script ...> opening tag so that
  // SvelteKit's bootstrap and module scripts are trusted by the browser.
  let html = await response.text();
  html = html.replace(/<script(\s|>)/g, `<script nonce="${nonce}"$1`);

  const headers = new Headers(response.headers);
  headers.set("content-security-policy", csp);

  return new Response(html, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

export const config: Config = {
  path: "/*",
};
