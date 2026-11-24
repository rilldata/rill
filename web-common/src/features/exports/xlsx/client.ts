import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import type { CanvasWorkbookExportRequest } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// Sends the workbook request to the runtime, which formats it as an .xlsx
// file (see runtime/server/workbook.go), and returns the file.
export async function postWorkbookRequest(
  client: RuntimeClient,
  request: CanvasWorkbookExportRequest,
): Promise<Blob> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  // The request follows a capture that can take a while, so the token may be close to expiring.
  const jwt = await client.getFreshJwt();
  if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

  const body = new TextEncoder().encode(request.toJsonString());
  let response: Response;
  try {
    response = await fetch(
      `${client.host}/v1/instances/${client.instanceId}/exports/workbook`,
      { method: "POST", headers, body },
    );
  } catch (e) {
    // fetch rejects with a bare "Failed to fetch" when no readable response comes back:
    // a dropped connection, or a proxy in front of the runtime refusing a large upload
    // with an error page that carries no CORS headers (nginx's default body limit is 1 MB).
    // Chart images are what make the request large, so name its size and how to shrink it.
    throw new Error(
      m.export_xlsx_upload_failed({ size: formatMemorySize(body.byteLength) }),
      { cause: e },
    );
  }
  if (!response.ok) {
    throw new Error(await errorMessage(response));
  }
  return response.blob();
}

async function errorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as {
      error?: string;
      message?: string;
    };
    const message = body.error ?? body.message;
    if (message) return message;
  } catch {
    // Not a JSON error body; fall through to the status.
  }
  return `${response.status} ${response.statusText}`.trim();
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  // Give the browser time to start the download before the URL is released.
  setTimeout(() => URL.revokeObjectURL(url), 10_000);
}
