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
  const jwt = client.getJwt();
  if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

  const response = await fetch(
    `${client.host}/v1/instances/${client.instanceId}/exports/workbook`,
    { method: "POST", headers, body: request.toJsonString() },
  );
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
