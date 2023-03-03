import { Batcher } from "@rilldata/web-common/runtime-client/batcher/Batcher";
import type { FetchWrapperOptions } from "@rilldata/web-local/lib/util/fetchWrapper";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";

let RuntimeUrl = "";
try {
  RuntimeUrl = RILL_RUNTIME_URL;
} catch (e) {
  // no-op
}

export const httpRequestQueue = new HttpRequestQueue(RuntimeUrl);
export const batcher = new Batcher(RuntimeUrl);

export const httpClient = async <T>(
  config: FetchWrapperOptions
): Promise<T> => {
  return batcher.add(config);
};

export default httpClient;
