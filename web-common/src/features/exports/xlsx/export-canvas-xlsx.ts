import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getFontEmbedCSS } from "html-to-image";
import { get } from "svelte/store";
import { buildExportFilename } from "../filename";
import {
  findExportView,
  measureCanvasBlocks,
  type MeasuredBlock,
} from "../pdf/capture";
import { prepareCanvasForCapture } from "../pdf/settle";
import { needsCanvasWarmup, rasterizeNode } from "../rasterize";
import { downloadBlob, postWorkbookRequest } from "./client";
import { buildWorkbookRequest, needsImage } from "./collect";
import { buildWorkbookHeader } from "./header";
import type { ExportCanvasXlsxOptions } from "./types";

// Orchestrates a canvas-to-Excel export: mount the off-screen export render
// (via exportMode, shared with the PDF export), read every component's data
// from its export payload, snapshot the components that need an image, send
// the workbook request to the runtime and download the file it returns.
// Stateless; the caller owns UI state (loading flag, notifications).
export async function exportCanvasXlsx(
  opts: ExportCanvasXlsxOptions,
): Promise<void> {
  const { canvasEntity } = getCanvasStore(opts.canvasName, opts.instanceId);

  canvasEntity.exportAllTabs.set(opts.allTabs);
  canvasEntity.exportMode.set(true);

  try {
    opts.onProgress?.({ phase: "preparing", ratio: 0 });
    await prepareCanvasForCapture(canvasEntity, queryClient, {
      instanceId: opts.instanceId,
      timeoutMs: opts.timeoutMs,
    });
    opts.onProgress?.({ phase: "preparing", ratio: 1 });

    const { exportView, rowContainer } = findExportView(
      opts.instanceId,
      opts.canvasName,
    );
    const blocks = measureCanvasBlocks(rowContainer).filter(
      (block) => !block.isTabLabel,
    );
    if (!blocks.length) {
      throw new Error("Nothing to export on this canvas.");
    }

    const images = await snapshotBlocks(exportView, blocks, canvasEntity, opts);

    const header = buildWorkbookHeader(
      canvasEntity,
      opts.canvasName,
      opts.includeFilters,
    );
    // UTC, e.g. "2026-06-19 02:20 UTC".
    const generatedAt = `${new Date().toISOString().replace("T", " ").slice(0, 16)} UTC`;
    const request = await buildWorkbookRequest({
      canvasName: opts.canvasName,
      canvasEntity,
      blocks,
      images,
      chartMode: opts.chartMode,
      header,
      generatedAt,
      dashboardUrl: opts.dashboardUrl ?? window.location.href,
      timeoutMs: opts.timeoutMs,
    });

    opts.onProgress?.({ phase: "assembling", ratio: 0 });
    const blob = await postWorkbookRequest(opts.client, request);
    downloadBlob(blob, buildExportFilename(header.title, "xlsx"));
    opts.onProgress?.({ phase: "assembling", ratio: 1 });
  } finally {
    canvasEntity.exportMode.set(false);
  }
}

// Rasterizes the blocks that are exported as images (see needsImage) to PNG.
// A block whose capture fails is left out and becomes a note in the workbook.
async function snapshotBlocks(
  exportView: HTMLElement,
  blocks: MeasuredBlock[],
  canvasEntity: ReturnType<typeof getCanvasStore>["canvasEntity"],
  opts: ExportCanvasXlsxOptions,
): Promise<Map<string, Uint8Array<ArrayBuffer>>> {
  const components = get(canvasEntity.componentsStore);
  const targets = blocks.filter((block) => {
    const component = components.get(block.id);
    return (
      !!component &&
      needsImage(component, get(component.exportPayload), opts.chartMode)
    );
  });

  const images = new Map<string, Uint8Array<ArrayBuffer>>();
  if (!targets.length) {
    opts.onProgress?.({ phase: "capturing", ratio: 1 });
    return images;
  }

  const backgroundColor = getComputedStyle(exportView).backgroundColor;
  // See captureCanvasBlocks for why these are resolved once per export.
  const warmUpCanvas = await needsCanvasWarmup();
  const fontEmbedCSS = await getFontEmbedCSS(exportView);

  let done = 0;
  for (const target of targets) {
    try {
      const dataUrl = await rasterizeNode(target.element, {
        backgroundColor,
        fontEmbedCSS,
        warmUpCanvas,
        format: "png",
      });
      images.set(target.id, dataUrlBytes(dataUrl));
    } catch (e) {
      console.warn(`Failed to capture canvas block "${target.id}"`, e);
    }
    done += 1;
    opts.onProgress?.({ phase: "capturing", ratio: done / targets.length });
  }
  return images;
}

function dataUrlBytes(dataUrl: string): Uint8Array<ArrayBuffer> {
  const base64 = dataUrl.slice(dataUrl.indexOf(",") + 1);
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}
