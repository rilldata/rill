/**
 * Decoding for the `open` embed param.
 * Emails from reports and alerts created in an embed link to the embedder's open page with a `rill_open` param.
 * The embedder forwards it to `GetIFrame`, which passes it on to the iframe as `open`.
 * The format mirrors `embedopen` in the runtime: unpadded base64url of a JSON object.
 */

export type EmbedOpenKind = "report" | "alert";

export type EmbedOpenPayload = {
  kind: EmbedOpenKind;
  name: string;
  executionTime?: string;
};

/**
 * Decodes an `open` param.
 * Returns null if it is malformed.
 */
export function decodeEmbedOpen(blob: string): EmbedOpenPayload | null {
  let raw: unknown;
  try {
    const base64 = blob.replace(/-/g, "+").replace(/_/g, "/");
    const binary = atob(base64.padEnd(Math.ceil(base64.length / 4) * 4, "="));
    const json = new TextDecoder().decode(
      Uint8Array.from(binary, (c) => c.charCodeAt(0)),
    );
    raw = JSON.parse(json);
  } catch {
    return null;
  }
  if (typeof raw !== "object" || raw === null) return null;

  const { kind, name, execution_time } = raw as Record<string, unknown>;
  if (kind !== "report" && kind !== "alert") return null;
  if (typeof name !== "string" || !name) return null;
  if (execution_time !== undefined && execution_time !== "") {
    if (
      typeof execution_time !== "string" ||
      Number.isNaN(Date.parse(execution_time))
    ) {
      return null;
    }
    return { kind, name, executionTime: execution_time };
  }
  return { kind, name };
}

/**
 * Returns the embed route that opens the report or alert in the payload.
 */
export function getEmbedOpenPath({
  kind,
  name,
  executionTime,
}: EmbedOpenPayload): string {
  const path = `/-/embed/${kind === "report" ? "reports" : "alerts"}/${encodeURIComponent(name)}/open`;
  if (!executionTime) return path;
  return `${path}?${new URLSearchParams({ execution_time: executionTime }).toString()}`;
}
