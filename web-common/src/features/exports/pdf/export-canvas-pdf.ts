import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { assemblePdf } from "./assemble";
import { captureCanvasBlocks } from "./capture";
import { buildPdfFilename } from "./filename";
import { paginate } from "./layout";
import { prepareCanvasForCapture } from "./settle";
import {
  DEFAULT_PDF_ORIENTATION,
  DEFAULT_PDF_PAGE_FORMAT,
  type ExportCanvasPdfOptions,
} from "./types";

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

  try {
    opts.onProgress?.({ phase: "preparing", ratio: 0 });
    await prepareCanvasForCapture(canvasEntity, queryClient, {
      instanceId: opts.instanceId,
      timeoutMs: opts.timeoutMs,
    });
    opts.onProgress?.({ phase: "preparing", ratio: 1 });

    const { blocks, contentWidthPx, backgroundColor } =
      await captureCanvasBlocks({
        includeFilters: opts.includeFilters,
        onProgress: (ratio) => opts.onProgress?.({ phase: "capturing", ratio }),
      });

    if (!blocks.length) {
      throw new Error("Nothing to export on this canvas.");
    }

    opts.onProgress?.({ phase: "assembling", ratio: 0 });
    const pagination = paginate(blocks, {
      contentWidthPx,
      format: DEFAULT_PDF_PAGE_FORMAT,
      orientation: DEFAULT_PDF_ORIENTATION,
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
    if (scrollContainer) scrollContainer.scrollTop = scrollTop;
  }
}
