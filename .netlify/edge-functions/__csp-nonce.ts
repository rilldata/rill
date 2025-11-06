// @ts-ignore cannot find module
import type { Config, Context } from "netlify:edge";
// @ts-ignore cannot find module
import { csp } from "https://deno.land/x/csp_nonce_html_transformer@v2.3.0/src/index-embedded-wasm.ts";

// Using `import ... with ...` syntax directly fails due to the node 18 type-checking we're running on this file,
// but this syntax works fine in deno 1.46.3 and 2.x which is what the functions are bundled and run with.
// We're able to sneak by the node syntax issues by using this `await import(...)` syntax instead of a direct import statement.
// @ts-ignore top-level await
const { default: inputs } = await import("./__csp-nonce-inputs.json", {
  // @ts-ignore 'with' syntax
  with: { type: "json" },
});

type Params = {
  reportOnly: boolean;
  reportUri?: string;
  unsafeEval: boolean;
  path: string | string[];
  excludedPath: string[];
  distribution?: string;
  strictDynamic?: boolean;
  unsafeInline?: boolean;
  self?: boolean;
  https?: boolean;
  http?: boolean;
};
const params = inputs as Params;
params.reportUri = params.reportUri || "/.netlify/functions/__csp-violations";
// @ts-ignore Netlify
params.distribution = Netlify.env.get("CSP_NONCE_DISTRIBUTION");

params.strictDynamic = params.strictDynamic ?? true;
params.unsafeInline = params.unsafeInline ?? true;
params.self = params.self ?? true;
params.https = true;
params.http = true;

const handler = async (_request: Request, context: Context) => {
  try {
    const response = await context.next();
    // for debugging which routes use this edge function
    response.headers.set("x-debug-csp-nonce", "invoked");
    return csp(response, params);
  } catch {
    /*
      We catch all the throws and return undefined
      The reason we do this is because returning undefined
      will cause the next edge function in the chain to be
      executed.
      This is equivalent to setting the Edge Function's 
      `config.onError` property to "bypass", but is handled 
      completely by the Edge Function instead of by something else.
    */
    return void 0;
  }
};

// Top 50 most common extensions (minus .html and .htm) according to Humio
const excludedExtensions = [
  "aspx",
  "avif",
  "babylon",
  "bak",
  "cgi",
  "com",
  "css",
  "ds",
  "env",
  "gif",
  "gz",
  "ico",
  "ini",
  "jpeg",
  "jpg",
  "js",
  "json",
  "jsp",
  "log",
  "m4a",
  "map",
  "md",
  "mjs",
  "mp3",
  "mp4",
  "ogg",
  "otf",
  "pdf",
  "php",
  "png",
  "rar",
  "sh",
  "sql",
  "svg",
  "ttf",
  "txt",
  "wasm",
  "wav",
  "webm",
  "webmanifest",
  "webp",
  "woff",
  "woff2",
  "xml",
  "xsd",
  "yaml",
  "yml",
  "zip",
];

export const config: Config = {
  path: params.path,
  excludedPath: ["/.netlify*", `**/*.(${excludedExtensions.join("|")})`]
    .concat(params.excludedPath)
    .filter(Boolean),
  handler,
  onError: "bypass",
  method: "GET",
};

export default handler;
