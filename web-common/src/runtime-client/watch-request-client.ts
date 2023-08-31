import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";
import { get, Unsubscriber } from "svelte/store";

type WatchResponse =
  | V1WatchFilesResponse
  | V1WatchResourcesResponse
  | V1WatchLogsResponse;

type StreamingFetchResponse<Res extends WatchResponse> = {
  result?: Res;
  error?: { code: number; message: string };
};

export class WatchRequestClient<Res extends WatchResponse> {
  private controller: AbortController;

  public constructor(
    private readonly getUrl: (runtime: Runtime) => string,
    private readonly onResponse: (res: Res) => void | Promise<void>,
    private readonly onReconnect: () => void | Promise<void>
  ) {}

  public start(): Unsubscriber {
    const unsubscribe = runtime.subscribe((runtimeState) => {
      this.controller?.abort();
      if (!runtimeState?.host || !runtimeState?.instanceId) return;

      this.maintainConnection();
    });

    return () => {
      this.controller?.abort();
      unsubscribe();
    };
  }

  private async maintainConnection() {
    let firstRun = true;
    // Maintain the controller here to make sure we check `aborted` for the correct one.
    // Checking for `this.controller` might lead to edge cases where it has a newer controller after `runtime` changed.
    let controller = new AbortController();

    const url = this.getUrl(get(runtime));

    while (!controller.signal.aborted) {
      if (!firstRun) {
        this.onReconnect();
        // safeguard to cancel the request if not already cancelled
        controller.abort();
        controller = new AbortController();
      }
      firstRun = false;

      this.controller = controller;
      try {
        const stream = streamingFetchWrapper<StreamingFetchResponse<Res>>(
          url,
          "GET",
          undefined,
          controller.signal
        );

        for await (const res of stream) {
          if (res.error) throw new Error(res.error.message);
          else if (res.result) this.onResponse(res.result);
        }
      } catch (err) {
        console.log(err);
        // TODO: make this smarter
        await asyncWait(2000);
      }
    }
  }
}
