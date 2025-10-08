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

type: theme

# Themes allow you to customize the appearance of Explore and Canvas dashboards
# with support for both light and dark modes. Use CSS to define custom colors,
# surfaces, and data visualization palettes (sequential, diverging, qualitative).

css: |
  :root {
    /* Primary brand color - used for lines, bars, and key UI elements */
    --primary: oklch(0.55 0.25 305);
    
    /* Secondary brand color - used for accents and secondary elements */
    --secondary: oklch(0.70 0.21 50);
    
    /* Surface/background color */
    --surface: oklch(0.98 0.01 85);
    
    /* Sequential palette (9 colors) */
    --color-sequential-1: oklch(0.95 0.03 320);
    --color-sequential-2: oklch(0.88 0.06 320);
    --color-sequential-3: oklch(0.80 0.10 320);
    --color-sequential-4: oklch(0.70 0.15 320);
    --color-sequential-5: oklch(0.60 0.20 320);
    --color-sequential-6: oklch(0.50 0.23 320);
    --color-sequential-7: oklch(0.42 0.20 320);
    --color-sequential-8: oklch(0.35 0.17 320);
    --color-sequential-9: oklch(0.28 0.14 320);
    
    /* Diverging palette (11 colors) */
    --color-diverging-1: oklch(0.60 0.24 30);
    --color-diverging-2: oklch(0.70 0.20 40);
    --color-diverging-3: oklch(0.80 0.15 50);
    --color-diverging-4: oklch(0.88 0.10 70);
    --color-diverging-5: oklch(0.94 0.05 90);
    --color-diverging-6: oklch(0.88 0.08 280);
    --color-diverging-7: oklch(0.78 0.12 290);
    --color-diverging-8: oklch(0.68 0.16 300);
    --color-diverging-9: oklch(0.58 0.20 310);
    --color-diverging-10: oklch(0.48 0.22 315);
    --color-diverging-11: oklch(0.38 0.20 320);
    
    /* Qualitative palette (12 colors) */
    --color-qualitative-1: oklch(0.70 0.20 340);
    --color-qualitative-2: oklch(0.75 0.22 60);
    --color-qualitative-3: oklch(0.65 0.19 280);
    --color-qualitative-4: oklch(0.72 0.18 160);
    --color-qualitative-5: oklch(0.78 0.20 110);
    --color-qualitative-6: oklch(0.80 0.16 40);
    --color-qualitative-7: oklch(0.68 0.21 10);
    --color-qualitative-8: oklch(0.60 0.17 240);
    --color-qualitative-9: oklch(0.65 0.19 180);
    --color-qualitative-10: oklch(0.73 0.18 90);
    --color-qualitative-11: oklch(0.66 0.20 310);
    --color-qualitative-12: oklch(0.77 0.17 130);
  }
  
  .dark {
    /* Primary brand color - dark mode (lighter for visibility) */
    --primary: oklch(0.70 0.20 305);
    
    /* Secondary brand color - dark mode (lighter for visibility) */
    --secondary: oklch(0.75 0.18 50);
    
    /* Surface/background color - dark mode */
    --surface: oklch(0.18 0.01 264);
    
    /* Sequential palette - Purple/Magenta (inverted for dark mode) */
    --color-sequential-1: oklch(0.40 0.17 320);
    --color-sequential-2: oklch(0.48 0.20 320);
    --color-sequential-3: oklch(0.56 0.23 320);
    --color-sequential-4: oklch(0.64 0.20 320);
    --color-sequential-5: oklch(0.70 0.17 320);
    --color-sequential-6: oklch(0.76 0.14 320);
    --color-sequential-7: oklch(0.82 0.11 320);
    --color-sequential-8: oklch(0.88 0.07 320);
    --color-sequential-9: oklch(0.92 0.04 320);
    
    /* Diverging palette - Orange to Purple (adjusted for dark) */
    --color-diverging-1: oklch(0.45 0.22 320);
    --color-diverging-2: oklch(0.52 0.24 315);
    --color-diverging-3: oklch(0.60 0.22 310);
    --color-diverging-4: oklch(0.68 0.18 300);
    --color-diverging-5: oklch(0.76 0.14 290);
    --color-diverging-6: oklch(0.82 0.10 280);
    --color-diverging-7: oklch(0.76 0.12 90);
    --color-diverging-8: oklch(0.68 0.16 70);
    --color-diverging-9: oklch(0.60 0.18 50);
    --color-diverging-10: oklch(0.52 0.20 40);
    --color-diverging-11: oklch(0.45 0.22 30);
    
    /* Qualitative palette - Vibrant colors (enhanced for dark mode) */
    --color-qualitative-1: oklch(0.68 0.22 340);
    --color-qualitative-2: oklch(0.73 0.24 60);
    --color-qualitative-3: oklch(0.64 0.21 280);
    --color-qualitative-4: oklch(0.70 0.20 160);
    --color-qualitative-5: oklch(0.75 0.22 110);
    --color-qualitative-6: oklch(0.77 0.18 40);
    --color-qualitative-7: oklch(0.66 0.23 10);
    --color-qualitative-8: oklch(0.62 0.19 240);
    --color-qualitative-9: oklch(0.67 0.21 180);
    --color-qualitative-10: oklch(0.71 0.20 90);
    --color-qualitative-11: oklch(0.64 0.22 310);
    --color-qualitative-12: oklch(0.74 0.19 130);
  }
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
