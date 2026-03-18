import type { Context } from "@netlify/edge-functions";

// Injects a per-request CSP nonce into HTML responses and sets the
// Content-Security-Policy header dynamically. This replaces the static
// Content-Security-Policy entries in netlify.toml, which cannot include
// nonces because Netlify header rules are evaluated at build time.
//
// With 'strict-dynamic': scripts loaded by a nonced script inherit trust, so
// third-party loaders (Pylon, Pusher) that dynamically inject child scripts
// work without needing their child scripts individually allowlisted.
// Domain allowlists are kept for backwards compatibility with older browsers
// that do not support 'strict-dynamic'.
export default async (
  request: Request,
  context: Context,
): Promise<Response> => {
  const response = await context.next();

  // Only process HTML documents; pass other assets through unchanged.
  if (!response.headers.get("content-type")?.includes("text/html")) {
    return response;
  }

  const nonceBytes = new Uint8Array(16);
  crypto.getRandomValues(nonceBytes);
  const nonce = btoa(String.fromCharCode(...nonceBytes));

  let body = await response.text();
  // Inject nonce onto every <script and <style opening tag.
  body = body.replace(/<script(?=[ >])/g, `<script nonce="${nonce}"`);
  body = body.replace(/<style(?=[ >])/g, `<style nonce="${nonce}"`);

  const url = new URL(request.url);
  const isEmbed = url.pathname.startsWith("/-/embed");
  const isShare = url.pathname.includes("/-/share");
  const isEmbeddable = isEmbed || isShare;

  // Embeddable routes allow framing from any HTTPS origin; the main app
  // restricts framing to same-origin only.
  const frameAncestors = isEmbeddable ? "https:" : "'self'";

  // ActiveCampaign (app-us1.com) is only loaded on the main app routes.
  const activeCampaign = isEmbeddable ? "" : " https://*.app-us1.com/";

  // Pylon CDN is used for styling on the main app; embeds keep styles tighter.
  const pylonStyles = isEmbeddable ? "" : " https://*.usepylon.com";

  const csp = [
    "default-src 'self'",
    // 'nonce-...' authorizes this document's own inline and external scripts.
    // 'strict-dynamic' propagates trust to scripts dynamically created by
    // those nonced scripts (e.g. Pylon/Pusher loaders injecting child scripts).
    // Domain allowlists below are fallbacks for browsers without strict-dynamic.
    // adding 'unsafe-inline' (ignored by browsers supporting nonces/hashes) to be backward compatible with older browsers.
    `script-src 'nonce-${nonce}' 'strict-dynamic' 'unsafe-inline' 'unsafe-eval'${activeCampaign} https://*.usepylon.com https://*.pusher.com`,
    // style-src keeps 'unsafe-inline' for now: runtime style injection from
    // CodeMirror and other libraries cannot be nonce-attributed. Revisit when
    // those libraries are audited.
    `style-src 'self' 'unsafe-inline'${pylonStyles}`,
    "img-src https: data: blob:",
    "frame-src 'self' https://www.youtube.com/ https://www.loom.com/ https://www.vimeo.com https://portal.withorb.com blob: data:",
    `frame-ancestors ${frameAncestors}`,
    "form-action 'self'",
    "object-src 'none'",
    "base-uri 'self'",
    "connect-src 'self' https://*.rilldata.com https://*.rilldata.io https://*.rilldata.in https://*.usepylon.com https://docs.google.com https://storage.googleapis.com https://cdn.prod.website-files.com https://*.stripe.com wss://*.pusher.com",
    "font-src 'self' https://fonts.gstatic.com https://*.usepylon.com",
  ].join("; ");

  const headers = new Headers(response.headers);
  headers.set("Content-Security-Policy", csp);

  return new Response(body, { status: response.status, headers });
};

export const config: Config = {
  path: "/*",
  rateLimit: {
    windowLimit: 300,
    windowSize: 60,
    aggregateBy: ["ip", "domain"],
  },
};
