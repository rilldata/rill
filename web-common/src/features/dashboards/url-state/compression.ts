import { protoBase64 } from "@bufbuild/protobuf";

const URL_SIZE_THRESHOLD = 2000;

export function shouldCompressParams(url: URL) {
  return url.toString().length > URL_SIZE_THRESHOLD;
}

export async function compressUrlParams(params: string) {
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

  return protoBase64.enc(compressed).replaceAll("+", "-").replaceAll("/", "_");
}

export async function decompressUrlParams(compressedParam: string) {
  const compressedParamAsUint8Array = protoBase64.dec(compressedParam);
  const compressedParamAsReadableStream = new ReadableStream({
    start: (controller) => {
      controller.enqueue(compressedParamAsUint8Array);
      controller.close();
    },
  });
  const decompressedParams = compressedParamAsReadableStream.pipeThrough(
    new DecompressionStream("gzip"),
  );
  const resp = new Response(decompressedParams);

  return resp.text();
}
