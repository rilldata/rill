import type {
  RuntimeServiceWatchFilesParams,
  RuntimeServiceWatchLogsParams,
  RuntimeServiceWatchResourcesParams,
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";

export type WatchRequest =
  | RuntimeServiceWatchFilesParams
  | RuntimeServiceWatchResourcesParams
  | RuntimeServiceWatchLogsParams;
export type WatchResponse =
  | V1WatchFilesResponse
  | V1WatchResourcesResponse
  | V1WatchLogsResponse;

type StreamingFetchResponse<Res extends WatchResponse> = {
  result?: Res;
  error?: { code: number; message: string };
};

export class WatchRequestClient<
  Req extends WatchRequest,
  Res extends WatchResponse
> {
  private controller: AbortController;
  private readonly url: string;
  private shouldCancel = false;

  public constructor(url: string, req: Req) {
    const urlObj = new URL(url);
    for (const p in req) {
      urlObj.searchParams.set(p, String(req[p]));
    }
    this.url = urlObj.toString();
  }

  public async *send(reconnected: () => void) {
    let firstRun = true;

    this.shouldCancel = false;
    while (!this.shouldCancel) {
      if (!firstRun) {
        reconnected();
      }
      firstRun = false;

      try {
        this.controller = new AbortController();
        const stream = streamingFetchWrapper<StreamingFetchResponse<Res>>(
          this.url,
          "GET",
          undefined,
          this.controller.signal
        );

        for await (const res of stream) {
          if (res.error) throw new Error(res.error.message);
          else if (res.result) yield res.result;
        }
      } catch (err) {
        console.log(err);
        // TODO: make this smarter
        await asyncWait(2000);
      }
    }
  }

  public cancel() {
    this.shouldCancel = true;
    this.controller?.abort();
  }
}
