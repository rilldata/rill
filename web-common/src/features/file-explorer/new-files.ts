import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import {
  ResourceKind,
  type UserFacingResourceKinds,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  runtimeServicePutFile,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function createResourceFile(
  kind: ResourceKind,
  baseResource?: V1Resource,
): Promise<string> {
  if (!(kind in ResourceKindMap)) {
    throw new Error(`Unknown resource kind: ${kind}`);
  }

  const newPath = getPathForNewResourceFile(kind, baseResource);
  const instanceId = get(runtime).instanceId;

  await runtimeServicePutFile(instanceId, {
    path: newPath,
    blob: generateBlobForNewResourceFile(kind, baseResource),
    create: true,
    createOnly: true,
  });

  // Return the path to the new file, so we can navigate the user to it
  return `/${newPath}`;
}

export function getPathForNewResourceFile(
  newKind: ResourceKind,
  baseResource?: V1Resource,
) {
  const allNames =
    newKind === ResourceKind.Source || newKind === ResourceKind.Model
      ? // sources and models share the name
        [
          ...fileArtifacts.getNamesForKind(ResourceKind.Source),
          ...fileArtifacts.getNamesForKind(ResourceKind.Model),
        ]
      : fileArtifacts.getNamesForKind(newKind);

  const { folderName, extension } = ResourceKindMap[newKind];
  const baseName = getBaseNameForNewResourceFile(newKind, baseResource);
  const newName = getName(baseName, allNames);

  return `${folderName}/${newName}${extension}`;
}

export const ResourceKindMap: Record<
  UserFacingResourceKinds,
  {
    folderName: string;
    baseName: string;
    extension: string;
  }
> = {
  [ResourceKind.Source]: {
    folderName: "sources",
    baseName: "source",
    extension: ".yaml",
  },
  [ResourceKind.Connector]: {
    folderName: "connectors",
    baseName: "connector",
    extension: ".yaml",
  },
  [ResourceKind.Model]: {
    folderName: "models",
    baseName: "model",
    extension: ".sql",
  },
  [ResourceKind.MetricsView]: {
    folderName: "metrics",
    baseName: "metrics_view",
    extension: ".yaml",
  },
  [ResourceKind.Explore]: {
    folderName: "dashboards",
    baseName: "explore",
    extension: ".yaml",
  },
  [ResourceKind.API]: {
    folderName: "apis",
    baseName: "api",
    extension: ".yaml",
  },
  [ResourceKind.Component]: {
    folderName: "components",
    baseName: "component",
    extension: ".yaml",
  },
  [ResourceKind.Canvas]: {
    folderName: "dashboards",
    baseName: "canvas",
    extension: ".yaml",
  },
  [ResourceKind.Theme]: {
    folderName: "themes",
    baseName: "theme",
    extension: ".yaml",
  },
  [ResourceKind.Report]: {
    folderName: "reports",
    baseName: "report",
    extension: ".yaml",
  },
  [ResourceKind.Alert]: {
    folderName: "alerts",
    baseName: "alert",
    extension: ".yaml",
  },
};

export function getBaseNameForNewResourceFile(
  newKind: ResourceKind,
  baseResource?: V1Resource,
) {
  switch (newKind) {
    case ResourceKind.Explore:
      return baseResource
        ? `${baseResource.meta!.name!.name}_explore`
        : ResourceKindMap[newKind].baseName;
    default:
      return ResourceKindMap[newKind].baseName;
  }
}

export function generateBlobForNewResourceFile(
  kind: ResourceKind,
  baseResource?: V1Resource,
) {
  switch (kind) {
    case ResourceKind.Connector:
      return ""; // This is constructed in the `features/connectors` directory
    case ResourceKind.Source:
      return ""; // This is constructed in the `features/sources/modal` directory
    case ResourceKind.Model:
      return `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/build/models

SELECT 'Hello, World!' AS Greeting`;
    case ResourceKind.MetricsView:
      return `# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: # Choose a model to underpin your metrics view
timeseries: # Choose a timestamp column (if any) from your model 

dimensions:
measures:
`;
    case ResourceKind.Explore:
      if (baseResource) {
        const metricsViewName = baseResource.meta!.name!.name;
        const metricsViewTitle =
          baseResource.metricsView?.state?.validSpec?.displayName;

        return `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

display_name: "${metricsViewTitle ? metricsViewTitle : metricsViewName} dashboard"
metrics_view: ${metricsViewName}

dimensions: '*'
measures: '*'
`;
      }
      return `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

display_name: "My metrics dashboard"
metrics_view: example_metrics_view # Choose a metrics view to underpin the dashboard

dimensions: '*'
measures: '*'
`;
    case ResourceKind.API:
      return `# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

type: api

metrics_sql: |
  select measure, dimension from metrics_view
`;
    case ResourceKind.Canvas:
      return `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/canvas-dashboards

type: canvas
display_name: "Canvas Dashboard"
defaults:
  time_range: PT24H
  comparison_mode: time
`;
    case ResourceKind.Theme:
      return `# Theme YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/themes
# This example shows a modern "Aurora" theme with indigo/purple gradients

type: theme

# Light mode colors
light:
  # Core brand colors
  primary: "#6366f1"     # Indigo for primary actions and emphasis
  secondary: "#8b5cf6"   # Purple for secondary elements
  
  # UI surface colors (optional - will use defaults if omitted)
  background: "#f8fafc"  # Soft gray background
  surface: "#ffffff"     # Clean white surfaces
  card: "#f1f5f9"        # Subtle card backgrounds
  
  # Qualitative palette (for categorical data - all 24 colors)
  color-qualitative-1: "#6366f1"   # Indigo
  color-qualitative-2: "#8b5cf6"   # Purple  
  color-qualitative-3: "#ec4899"   # Pink
  color-qualitative-4: "#06b6d4"   # Cyan
  color-qualitative-5: "#10b981"   # Emerald
  color-qualitative-6: "#f59e0b"   # Amber
  color-qualitative-7: "#3b82f6"   # Blue
  color-qualitative-8: "#a855f7"   # Violet
  color-qualitative-9: "#ef4444"   # Red
  color-qualitative-10: "#14b8a6"  # Teal
  color-qualitative-11: "#84cc16"  # Lime
  color-qualitative-12: "#f97316"  # Orange
  color-qualitative-13: "#d946ef"  # Fuchsia
  color-qualitative-14: "#eab308"  # Yellow
  color-qualitative-15: "#0ea5e9"  # Sky
  color-qualitative-16: "#a855f7"  # Purple alt
  color-qualitative-17: "#22c55e"  # Green
  color-qualitative-18: "#fb923c"  # Orange alt
  color-qualitative-19: "#f43f5e"  # Rose
  color-qualitative-20: "#6366f1"  # Indigo alt
  color-qualitative-21: "#2dd4bf"  # Teal alt
  color-qualitative-22: "#facc15"  # Yellow alt
  color-qualitative-23: "#c084fc"  # Violet alt
  color-qualitative-24: "#4ade80"  # Green alt
  
  # Sequential palette (for ordered data, light to dark - 9 colors)
  color-sequential-1: "#eef2ff"   # Lightest indigo
  color-sequential-2: "#e0e7ff"
  color-sequential-3: "#c7d2fe"
  color-sequential-4: "#a5b4fc"
  color-sequential-5: "#818cf8"
  color-sequential-6: "#6366f1"
  color-sequential-7: "#4f46e5"
  color-sequential-8: "#4338ca"
  color-sequential-9: "#3730a3"   # Darkest indigo
  
  # Diverging palette (for data with a meaningful midpoint - 11 colors)
  color-diverging-1: "#dc2626"    # Red (negative extreme)
  color-diverging-2: "#f87171"
  color-diverging-3: "#fca5a5"
  color-diverging-4: "#fecaca"
  color-diverging-5: "#fee2e2"
  color-diverging-6: "#f3f4f6"    # Neutral gray (midpoint)
  color-diverging-7: "#dbeafe"
  color-diverging-8: "#93c5fd"
  color-diverging-9: "#60a5fa"
  color-diverging-10: "#3b82f6"
  color-diverging-11: "#2563eb"   # Blue (positive extreme)

# Dark mode colors
dark:
  # Core brand colors (brighter for dark backgrounds)
  primary: "#818cf8"     # Lighter indigo for visibility
  secondary: "#a78bfa"   # Lighter purple
  
  # UI surface colors (optional)
  background: "#0f172a"  # Deep slate background
  surface: "#1e293b"     # Elevated surfaces
  card: "#334155"        # Card backgrounds
  
  # Qualitative palette (adjusted for dark mode visibility - all 24 colors)
  color-qualitative-1: "#818cf8"   # Indigo
  color-qualitative-2: "#a78bfa"   # Purple
  color-qualitative-3: "#f472b6"   # Pink
  color-qualitative-4: "#22d3ee"   # Cyan
  color-qualitative-5: "#34d399"   # Emerald
  color-qualitative-6: "#fbbf24"   # Amber
  color-qualitative-7: "#60a5fa"   # Blue
  color-qualitative-8: "#c084fc"   # Violet
  color-qualitative-9: "#f87171"   # Red
  color-qualitative-10: "#2dd4bf"  # Teal
  color-qualitative-11: "#a3e635"  # Lime
  color-qualitative-12: "#fb923c"  # Orange
  color-qualitative-13: "#e879f9"  # Fuchsia
  color-qualitative-14: "#facc15"  # Yellow
  color-qualitative-15: "#38bdf8"  # Sky
  color-qualitative-16: "#c084fc"  # Purple alt
  color-qualitative-17: "#4ade80"  # Green
  color-qualitative-18: "#fdba74"  # Orange alt
  color-qualitative-19: "#fb7185"  # Rose
  color-qualitative-20: "#818cf8"  # Indigo alt
  color-qualitative-21: "#5eead4"  # Teal alt
  color-qualitative-22: "#fde047"  # Yellow alt
  color-qualitative-23: "#d8b4fe"  # Violet alt
  color-qualitative-24: "#86efac"  # Green alt
  
  # Sequential palette (dark to light for dark mode - 9 colors)
  color-sequential-1: "#312e81"   # Darkest indigo
  color-sequential-2: "#3730a3"
  color-sequential-3: "#4338ca"
  color-sequential-4: "#4f46e5"
  color-sequential-5: "#6366f1"
  color-sequential-6: "#818cf8"
  color-sequential-7: "#a5b4fc"
  color-sequential-8: "#c7d2fe"
  color-sequential-9: "#e0e7ff"   # Lightest indigo
  
  # Diverging palette (adjusted for dark backgrounds - 11 colors)
  color-diverging-1: "#ef4444"    # Red (negative extreme)
  color-diverging-2: "#f87171"
  color-diverging-3: "#fca5a5"
  color-diverging-4: "#fecaca"
  color-diverging-5: "#fee2e2"
  color-diverging-6: "#475569"    # Neutral slate (midpoint)
  color-diverging-7: "#bfdbfe"
  color-diverging-8: "#93c5fd"
  color-diverging-9: "#60a5fa"
  color-diverging-10: "#3b82f6"
  color-diverging-11: "#2563eb"   # Blue (positive extreme)
`;
    case ResourceKind.Alert:
      return `# Alert YAML
# Reference documentation: TODO

type: alert

...
`;
    case ResourceKind.Report:
      return `# Report YAML
# Reference documentation: TODO

type: report

...
`;
    default:
      throw new Error(`Unknown resource kind: ${kind}`);
  }
}
