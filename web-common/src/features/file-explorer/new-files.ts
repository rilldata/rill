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
    --primary: hsl(274deg 76% 53%);
    
    /* Secondary brand color - used for accents and secondary elements */
    --secondary: hsl(25deg 100% 50%);
    
    /* Surface/background color */
    --surface: hsl(40deg 60% 97%);

    /* Background color */
    --background: hsl(40deg 60% 97%);
    
    /* Sequential palette (9 colors) */
    --color-sequential-1: hsl(289deg 72% 95%);
    --color-sequential-2: hsl(289deg 56% 87%);
    --color-sequential-3: hsl(289deg 55% 78%);
    --color-sequential-4: hsl(289deg 54% 67%);
    --color-sequential-5: hsl(289deg 53% 54%);
    --color-sequential-6: hsl(290deg 81% 38%);
    --color-sequential-7: hsl(291deg 97% 28%);
    --color-sequential-8: hsl(291deg 100% 22%);
    --color-sequential-9: hsl(290deg 100% 16%);
    
    /* Diverging palette (11 colors) */
    --color-diverging-1: hsl(5deg 99% 47%);
    --color-diverging-2: hsl(17deg 100% 58%);
    --color-diverging-3: hsl(24deg 100% 69%);
    --color-diverging-4: hsl(33deg 100% 78%);
    --color-diverging-5: hsl(44deg 78% 87%);
    --color-diverging-6: hsl(233deg 100% 90%);
    --color-diverging-7: hsl(249deg 100% 83%);
    --color-diverging-8: hsl(263deg 73% 71%);
    --color-diverging-9: hsl(277deg 58% 56%);
    --color-diverging-10: hsl(285deg 76% 39%);
    --color-diverging-11: hsl(290deg 100% 25%);
    
    /* Qualitative palette (12 colors) */
    --color-qualitative-1: hsl(317deg 78% 65%);
    --color-qualitative-2: hsl(30deg 100% 50%);
    --color-qualitative-3: hsl(240deg 99% 74%);
    --color-qualitative-4: hsl(158deg 100% 39%);
    --color-qualitative-5: hsl(60deg 100% 38%);
    --color-qualitative-6: hsl(18deg 100% 71%);
    --color-qualitative-7: hsl(345deg 98% 65%);
    --color-qualitative-8: hsl(202deg 100% 43%);
    --color-qualitative-9: hsl(170deg 100% 35%);
    --color-qualitative-10: hsl(45deg 100% 41%);
    --color-qualitative-11: hsl(276deg 77% 66%);
    --color-qualitative-12: hsl(87deg 54% 53%);
  }
  
  .dark {
    /* Primary brand color - dark mode (lighter for visibility) */
    --primary: hsl(270deg 100% 73%);
    
    /* Secondary brand color - dark mode (lighter for visibility) */
    --secondary: hsl(25deg 100% 59%);
    
    /* Surface/background color - dark mode */
    --surface: hsl(220deg 18% 7%);
    
    /* Sequential palette - Purple/Magenta (inverted for dark mode) */
    --color-sequential-1: hsl(290deg 69% 29%);
    --color-sequential-2: hsl(290deg 64% 39%);
    --color-sequential-3: hsl(290deg 61% 48%);
    --color-sequential-4: hsl(289deg 60% 60%);
    --color-sequential-5: hsl(289deg 62% 67%);
    --color-sequential-6: hsl(289deg 65% 74%);
    --color-sequential-7: hsl(289deg 70% 81%);
    --color-sequential-8: hsl(289deg 68% 87%);
    --color-sequential-9: hsl(289deg 57% 91%);
    
    /* Diverging palette - Orange to Purple (adjusted for dark) */
    --color-diverging-1: hsl(291deg 100% 31%);
    --color-diverging-2: hsl(285deg 76% 44%);
    --color-diverging-3: hsl(277deg 70% 59%);
    --color-diverging-4: hsl(264deg 86% 72%);
    --color-diverging-5: hsl(250deg 100% 81%);
    --color-diverging-6: hsl(235deg 100% 86%);
    --color-diverging-7: hsl(44deg 57% 56%);
    --color-diverging-8: hsl(37deg 100% 42%);
    --color-diverging-9: hsl(25deg 100% 41%);
    --color-diverging-10: hsl(12deg 100% 38%);
    --color-diverging-11: hsl(0deg 100% 34%);
    
    /* Qualitative palette - Vibrant colors (enhanced for dark mode) */
    --color-qualitative-1: hsl(316deg 79% 62%);
    --color-qualitative-2: hsl(27deg 100% 50%);
    --color-qualitative-3: hsl(242deg 100% 73%);
    --color-qualitative-4: hsl(155deg 100% 38%);
    --color-qualitative-5: hsl(59deg 100% 36%);
    --color-qualitative-6: hsl(18deg 100% 66%);
    --color-qualitative-7: hsl(343deg 98% 61%);
    --color-qualitative-8: hsl(203deg 100% 46%);
    --color-qualitative-9: hsl(169deg 100% 37%);
    --color-qualitative-10: hsl(43deg 100% 41%);
    --color-qualitative-11: hsl(277deg 81% 64%);
    --color-qualitative-12: hsl(83deg 72% 44%);
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
