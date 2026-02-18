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
  isMaterialized?: boolean; // has outputConnector (writes to external store)
  outputConnector?: string; // output connector name
  stageConnector?: string; // stage connector name
  changeMode?: string; // model change mode (e.g. "CHANGE_MODE_FULL", "CHANGE_MODE_APPEND")
  inputConnector?: string; // input connector name (e.g., "duckdb", "s3")
  resultTable?: string; // materialized result table name
  materialize?: boolean; // whether model materializes output
  refUpdate?: boolean; // refresh on upstream ref update
  timeoutSeconds?: number; // execution timeout
  retryAttempts?: number; // number of retry attempts on failure
  retryDelaySeconds?: number; // delay between retries
  retryExponentialBackoff?: boolean; // whether retries use exponential backoff
  retryIfErrorMatches?: string[]; // error patterns that trigger retry
  executionDurationMs?: string; // latest execution duration in ms
  testCount?: number; // number of tests defined
  testErrors?: string[]; // test failure messages (empty = all pass)

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
  exploreMeasuresCount?: number; // explicit measures selected
  exploreDimensionsCount?: number; // explicit dimensions selected
  exploreMeasuresAll?: boolean; // uses selector with all: true
  exploreDimensionsAll?: boolean; // uses selector with all: true

  // Canvas metadata
  componentCount?: number;
  rowCount?: number;

  // Security
  hasSecurityRules?: boolean;

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
export type ResourceStatusFilterValue = "ok" | "pending" | "errored";
export type ResourceStatusFilter = ResourceStatusFilterValue[];
