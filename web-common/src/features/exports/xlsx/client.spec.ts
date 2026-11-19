import {
  CanvasWorkbookExportRequest,
  WorkbookBlock,
  WorkbookDashboardSheet,
  WorkbookImageBlock,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, describe, expect, it, vi } from "vitest";
import { postWorkbookRequest } from "./client";

describe("postWorkbookRequest", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("names the upload size when the request gets no response", async () => {
    // What a browser reports when a proxy refuses the upload without CORS headers.
    const networkError = new TypeError("Failed to fetch");
    vi.stubGlobal("fetch", () => Promise.reject(networkError));
    // 1.5 MiB of PNG bytes is 2 MiB once base64-encoded in the JSON body.
    const request = new CanvasWorkbookExportRequest({
      canvas: "sales",
      dashboardSheets: [
        new WorkbookDashboardSheet({
          blocks: [
            new WorkbookBlock({
              content: {
                case: "image",
                value: new WorkbookImageBlock({
                  png: new Uint8Array(1.5 * 1024 * 1024),
                }),
              },
            }),
          ],
        }),
      ],
    });
    const client = new RuntimeClient({
      host: "http://localhost:9009",
      instanceId: "default",
    });

    const error = await postWorkbookRequest(client, request).catch(
      (e: unknown) => e,
    );

    expect(error).toBeInstanceOf(Error);
    expect((error as Error).message).toContain("(2.0MB)");
    expect((error as Error).cause).toBe(networkError);
  });
});
