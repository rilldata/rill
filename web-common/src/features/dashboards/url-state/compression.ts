import { protoBase64 } from "@bufbuild/protobuf";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

const URL_SIZE_THRESHOLD = 2000;

export async function compressUrlParams(url: URL) {
  const params = url.searchParams.toString();
  if (url.toString().length <= URL_SIZE_THRESHOLD) return params;

  const resp = new Response(
    new ReadableStream({
      start: (controller) => {
        controller.enqueue(new TextEncoder().encode(params));
        controller.close();
      },
    }).pipeThrough(new CompressionStream("gzip")),
  );
  const compressed = new Uint8Array(await resp.arrayBuffer());

  return protoBase64.enc(compressed);
}

export async function decompressUrlParams(searchParams: URLSearchParams) {
  if (!searchParams.has(ExploreStateURLParams.GzippedParams))
    return searchParams;

  const compressed = protoBase64.dec(
    searchParams.get(ExploreStateURLParams.GzippedParams)!,
  );
  const resp = new Response(
    new ReadableStream({
      start: (controller) => {
        controller.enqueue(compressed);
        controller.close();
      },
    }).pipeThrough(new DecompressionStream("gzip")),
  );
  const newUrl = new URL("http://localhost");
  newUrl.search = await resp.text();
  return newUrl.searchParams;
}
