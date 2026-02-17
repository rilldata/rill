import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

/**
 * Dimension info for MetricsView
 */
export interface DimensionInfo {
  name: string;
  displayName?: string;
  description?: string;
  type?: string; // CATEGORICAL | TIME | GEOSPATIAL
  column?: string;
  expression?: string;
}

/**
 * Measure info for MetricsView
 */
export interface MeasureInfo {
  name: string;
  displayName?: string;
  description?: string;
  expression?: string;
  type?: string; // SIMPLE | DERIVED | TIME_COMPARISON
}

/**
 * Rich metadata extracted from resources for Describe modal display.
 * All fields are optional since not all resource types have all metadata.
 */
export interface ResourceMetadata {
  // Model/Source metadata
  connector?: string; // inputConnector name
  sourcePath?: string; // file path for file-based sources
  incremental?: boolean;
  partitioned?: boolean;
  hasSchedule?: boolean;
  scheduleDescription?: string; // "cron: 0 * * * *" or "every 3600s"
  lastRefreshedOn?: string; // ISO timestamp of last execution
  isSqlModel?: boolean; // true if model defined via SQL file
  sqlQuery?: string; // SQL query for models
  testCount?: number; // number of tests defined

  // Dashboard metadata
  theme?: string; // theme name (not embedded)

  // MetricsView metadata
  metricsTable?: string;
  metricsModel?: string;
  timeDimension?: string;
  dimensions?: DimensionInfo[];
  measures?: MeasureInfo[];

  // Explore metadata
  metricsViewName?: string;

  // Canvas metadata
  componentCount?: number;
  rowCount?: number;

  // Consumer counts (for any resource)
  alertCount?: number;
  apiCount?: number;
}

export interface ResourceNodeData extends Record<string, unknown> {
  resource: V1Resource;
  kind?: ResourceKind;
  label: string;
  // transient UI flag to emphasize nodes along the traced path
  routeHighlighted?: boolean;
  // true when the node represents the seeded/root resource for the graph card
  isRoot?: boolean;
  // Rich metadata for badge indicators
  metadata?: ResourceMetadata;
  // Whether to show the "..." actions button on this node
  showNodeActions?: boolean;
}

/**
 * Filter values for resource status in the graph view.
 * Empty array means "all" (no filter applied).
 * - "pending": Resources with non-idle reconcile status
 * - "errored": Resources with reconcile errors
 */
export type ResourceStatusFilterValue = "pending" | "errored";
export type ResourceStatusFilter = ResourceStatusFilterValue[];
