export type PdfPageFormat = "a4" | "letter";

// "auto" resolves to landscape for wide canvases and portrait otherwise.
export type PdfOrientation = "portrait" | "landscape" | "auto";

export const DEFAULT_PDF_PAGE_FORMAT: PdfPageFormat = "a4";
export const DEFAULT_PDF_ORIENTATION: PdfOrientation = "auto";

// Resolved orientation (after "auto" has been decided).
export type ResolvedOrientation = "portrait" | "landscape";

export interface ExportProgress {
  phase: "preparing" | "capturing" | "assembling";
  // 0..1 within the current phase.
  ratio: number;
}

// The options collected by the shared ExportDashboardForm. Surface-specific
// orchestrators (canvas, explore) receive these plus their own identifiers.
export interface PdfExportRunOptions {
  includeFilters: boolean;
  onProgress?: (progress: ExportProgress) => void;
}

export interface ExportCanvasPdfOptions {
  canvasName: string;
  instanceId: string;
  includeFilters: boolean;
  timeoutMs?: number;
  onProgress?: (progress: ExportProgress) => void;
}

// Sentinel rowIndex/id for the filter bar block, which always renders first.
export const FILTER_BAR_ID = "__filter_bar__";
export const FILTER_BAR_ROW_INDEX = -1;

// A rasterized block (the filter bar or a single component) plus its position
// and size in the canvas content area, measured in CSS pixels.
export interface CapturedBlock {
  id: string;
  dataUrl: string;
  // Position relative to the content area's top-left, in CSS pixels.
  xPx: number;
  yPx: number;
  widthPx: number;
  heightPx: number;
  // Components sharing a rowIndex are laid out on the same canvas row.
  rowIndex: number;
}
