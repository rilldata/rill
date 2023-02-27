import type { FetchWrapperOptions } from "@rilldata/web-local/lib/util/fetchWrapper";
import { get } from "svelte/store";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { runtime } from "./runtime-store";

/**
 * Runtime base URL
 *  Local
 *    In dev & prod: http://localhost:9009
 *  Cloud
 *    In dev: http://localhost:9009
 *    In prod: https://{region}.runtime.rilldata.com
 */

const httpRequestQueues = new Map<string, HttpRequestQueue>();

export const httpClient = async <T>(
  config: FetchWrapperOptions
): Promise<T> => {
  // naive request interceptor
  const host = get(runtime).host;
  const interceptedConfig = { ...config, baseUrl: host };

  const httpRequestQueue = getHttpRequestQueueForHost(host);
  return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
};

export function getHttpRequestQueueForHost(host: string): HttpRequestQueue {
  if (!httpRequestQueues.has(host)) {
    httpRequestQueues.set(host, new HttpRequestQueue(host));
  }

  return httpRequestQueues.get(host);
}

export default httpClient;
