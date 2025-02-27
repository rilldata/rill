import { protoBase64 } from "@bufbuild/protobuf";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

const URL_SIZE_THRESHOLD = 2000;

export async function compressUrlParams(url: URL) {
  const params = url.searchParams.toString();
  if (url.toString().length <= URL_SIZE_THRESHOLD) return params;

  const paramsAsUint8Array = new TextEncoder().encode(params);
  const paramsAsReadableStream = new ReadableStream({
    start: (controller) => {
      controller.enqueue(paramsAsUint8Array);
      controller.close();
    },
  });
  const gzippedParams = paramsAsReadableStream.pipeThrough(
    new CompressionStream("gzip"),
  );
  const resp = new Response(gzippedParams);
  const compressed = new Uint8Array(await resp.arrayBuffer());

  const newUrl = new URL("http://localhost");
  newUrl.searchParams.set(
    ExploreStateURLParams.GzippedParams,
    protoBase64.enc(compressed).replaceAll("+", "-").replaceAll("/", "_"),
  );
  return newUrl.searchParams;
}

export async function decompressUrlParams(searchParams: URLSearchParams) {
  if (!searchParams.has(ExploreStateURLParams.GzippedParams))
    return searchParams;

  const compressedParam = protoBase64.dec(
    searchParams.get(ExploreStateURLParams.GzippedParams)!,
  );
  const compressedParamAsReadableStream = new ReadableStream({
    start: (controller) => {
      controller.enqueue(compressedParam);
      controller.close();
    },
  });
  const decompressedParams = compressedParamAsReadableStream.pipeThrough(
    new DecompressionStream("gzip"),
  );
  const resp = new Response(decompressedParams);

  const newUrl = new URL("http://localhost");
  newUrl.search = await resp.text();
  return newUrl.searchParams;
}
