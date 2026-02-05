import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

/**
 * Rich metadata extracted from resources for badge display.
 * All fields are optional since not all resource types have all metadata.
 */
export interface ResourceMetadata {
  // Model/Source metadata
  connector?: string; // inputConnector name
  incremental?: boolean;
  partitioned?: boolean;
  hasSchedule?: boolean;
  scheduleDescription?: string; // "cron: 0 * * * *" or "every 3600s"
  retryAttempts?: number;
  isSqlModel?: boolean; // true if model defined via SQL file
  isIntermediate?: boolean; // true if model has both upstream and downstream deps

  // Dashboard metadata
  theme?: string; // theme name (not embedded)

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
