import { protoBase64 } from "@bufbuild/protobuf";
import { gzipSync, gunzipSync } from "fflate";

const URL_SIZE_THRESHOLD = 2000;

export function shouldCompressParams(url: URL) {
  return url.toString().length > URL_SIZE_THRESHOLD;
}

export function compressUrlParams(params: string) {
  const paramsAsUint8Array = new TextEncoder().encode(params);
  const compressed = gzipSync(paramsAsUint8Array, {
    // Since we are not compression a file disable modified time to avoid output changing based on time.
    mtime: 0,
  });
  return protoBase64.enc(compressed).replaceAll("+", "-").replaceAll("/", "_");
}

export function decompressUrlParams(compressedParam: string) {
  const compressedParamAsUint8Array = protoBase64.dec(compressedParam);
  const decompressedParams = gunzipSync(compressedParamAsUint8Array);
  return String.fromCharCode(...decompressedParams);
}
