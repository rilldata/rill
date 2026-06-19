import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { assemblePdf } from "./assemble";
import { captureCanvasBlocks } from "./capture";
import { buildPdfFilename } from "./filename";
import { expandTablesForCapture } from "./expand-tables";
import { paginate } from "./layout";
import { prepareCanvasForCapture } from "./settle";
import type { ExportCanvasPdfOptions } from "./types";

// Orchestrates a client-side canvas-to-PDF export: force-render the canvas,
// rasterize the filter bar + each component, paginate to mirror the on-screen
// layout, then assemble and download the PDF. Stateless; the caller owns UI
// state (loading flag, notifications).
export async function exportCanvasPdf(
  opts: ExportCanvasPdfOptions,
): Promise<void> {
  const { canvasEntity } = getCanvasStore(opts.canvasName, opts.instanceId);

  const scrollContainer = document.querySelector<HTMLElement>(
    "#canvas-scroll-container",
  );
  const scrollTop = scrollContainer?.scrollTop ?? 0;

  let restoreTables: (() => void) | undefined;
  try {
    opts.onProgress?.({ phase: "preparing", ratio: 0 });
    await prepareCanvasForCapture(canvasEntity, queryClient, {
      timeoutMs: opts.timeoutMs,
    });
    opts.onProgress?.({ phase: "preparing", ratio: 1 });

    const rowContainer =
      scrollContainer?.querySelector<HTMLElement>(".row-container");
    const { restore, truncatedNotes } = expandTablesForCapture(
      rowContainer ?? document,
      opts.tableRowCap,
    );
    restoreTables = restore;
    // Let the virtualizer materialize the now-visible rows before capturing.
    await new Promise((r) => requestAnimationFrame(() => r(null)));
    await new Promise((r) => requestAnimationFrame(() => r(null)));

    const { blocks, contentWidthPx, backgroundColor } =
      await captureCanvasBlocks({
        includeFilters: opts.includeFilters,
        truncatedNotes,
        onProgress: (ratio) => opts.onProgress?.({ phase: "capturing", ratio }),
      });

    if (!blocks.length) {
      throw new Error("Nothing to export on this canvas.");
    }

    opts.onProgress?.({ phase: "assembling", ratio: 0 });
    const pagination = paginate(blocks, {
      contentWidthPx,
      format: opts.format,
      orientation: opts.orientation,
    });

    const title = get(canvasEntity.titleStore) || opts.canvasName;
    // UTC, e.g. "2026-06-19 02:20 UTC".
    const generatedAt = `${new Date().toISOString().replace("T", " ").slice(0, 16)} UTC`;
    await assemblePdf(pagination, {
      title,
      filename: buildPdfFilename(title),
      backgroundColor,
      generatedAt,
    });
    opts.onProgress?.({ phase: "assembling", ratio: 1 });
  } finally {
    restoreTables?.();
    if (scrollContainer) scrollContainer.scrollTop = scrollTop;
  }
}
