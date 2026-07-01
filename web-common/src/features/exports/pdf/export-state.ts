import { derived, writable, type Readable } from "svelte/store";

// Set of canvases (keyed by instance + canvas name) that are currently
// capturing a PDF export. Keyed rather than a single global boolean so that,
// when several canvases are mounted at once (e.g. embeds), only the exporting
// one mounts its capture header.
const activeExports = writable(new Set<string>());

function exportKey(instanceId: string, canvasName: string): string {
  return `${instanceId}:${canvasName}`;
}

// Marks a canvas as actively exporting (or clears it). Gates the off-screen
// capture header (see CanvasPdfExportHeader) so the header exists in the DOM
// solely during export. Playwright locators (and the accessibility tree) match
// elements regardless of CSS visibility, so an always-mounted, merely-hidden
// header would still duplicate the live filter bar's text and labels; only
// removing it from the DOM avoids that.
export function setCanvasPdfExportActive(
  instanceId: string,
  canvasName: string,
  active: boolean,
): void {
  const key = exportKey(instanceId, canvasName);
  activeExports.update((keys) => {
    const next = new Set(keys);
    if (active) next.add(key);
    else next.delete(key);
    return next;
  });
}

// True only while the given canvas is capturing a PDF export.
export function canvasPdfExportActive(
  instanceId: string,
  canvasName: string,
): Readable<boolean> {
  const key = exportKey(instanceId, canvasName);
  return derived(activeExports, (keys) => keys.has(key));
}
