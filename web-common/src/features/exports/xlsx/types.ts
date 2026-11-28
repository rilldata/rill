import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { ExportProgress } from "../pdf/types";

// How charts appear on the dashboard sheet: native Excel charts bound to the
// data sheets, or PNG snapshots of the rendered components. Data sheets are
// written in both modes.
export type XlsxChartMode = "native" | "image";

export interface ExportCanvasXlsxOptions {
  canvasName: string;
  instanceId: string;
  client: RuntimeClient;
  includeFilters: boolean;
  allTabs: boolean;
  chartMode: XlsxChartMode;
  timeoutMs?: number;
  // URL written in the workbook header. Defaults to the current page URL;
  // override it when the current URL should not be shared (e.g. it carries an auth token).
  dashboardUrl?: string;
  onProgress?: (progress: ExportProgress) => void;
}
