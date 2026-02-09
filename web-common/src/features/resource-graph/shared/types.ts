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
 * Rich metadata extracted from resources for inspector display.
 * All fields are optional since not all resource types have all metadata.
 */
export interface ResourceMetadata {
  // Model/Source metadata
  connector?: string; // inputConnector name
  outputConnector?: string; // output connector name
  stageConnector?: string; // staging connector name
  sourcePath?: string; // file path for file-based sources
  incremental?: boolean;
  partitioned?: boolean;
  partitionsWatermarkField?: string;
  partitionsConcurrencyLimit?: number;
  hasSchedule?: boolean;
  scheduleDescription?: string; // "cron: 0 * * * *" or "every 3600s"
  timeoutSeconds?: number;
  retryAttempts?: number;
  retryDelaySeconds?: number;
  retryExponentialBackoff?: boolean;
  retryYaml?: string; // YAML representation of retry config
  isSqlModel?: boolean; // true if model defined via SQL file
  sqlQuery?: string; // SQL query for models
  changeMode?: string; // RESET | MANUAL | PATCH
  testCount?: number; // number of tests defined
  testsYaml?: string; // YAML representation of tests
  scheduleYaml?: string; // YAML representation of schedule

  // Model state
  refreshedOn?: string;
  resultTable?: string;
  executionDurationMs?: string;
  partitionsHaveErrors?: boolean;

  // Dashboard metadata
  theme?: string; // theme name (not embedded)

  // MetricsView metadata
  metricsConnector?: string;
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
}

/**
 * Filter type for resource status in the graph view.
 * - "all": Show all resources
 * - "pending": Show only resources with non-idle reconcile status
 * - "errored": Show only resources with reconcile errors
 */
export type ResourceStatusFilter = "all" | "pending" | "errored";
