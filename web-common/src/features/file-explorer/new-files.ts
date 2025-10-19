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
# with support for both light and dark modes. All properties are optional - any
# properties not defined will fall back to Rill's default theme.

light:
  # Primary brand color - used for lines, bars, and key UI elements
  primary: hsl(180deg 65% 45%)
  
  # Secondary brand color - used for accents and secondary elements
  secondary: hsl(45deg 85% 55%)
  
  # Surface/background colors
  surface: hsl(180deg 30% 96%)
  background: hsl(180deg 30% 98%)
  
  # Sequential palette (9 colors) - for gradients and heatmaps (Teal theme)
  color-sequential-1: hsl(180deg 70% 95%)
  color-sequential-2: hsl(180deg 65% 85%)
  color-sequential-3: hsl(180deg 60% 75%)
  color-sequential-4: hsl(180deg 65% 65%)
  color-sequential-5: hsl(180deg 65% 55%)
  color-sequential-6: hsl(180deg 70% 45%)
  color-sequential-7: hsl(180deg 75% 35%)
  color-sequential-8: hsl(180deg 80% 25%)
  color-sequential-9: hsl(180deg 85% 18%)
  
  # Diverging palette (11 colors) - for positive/negative comparisons (Amber to Teal)
  color-diverging-1: hsl(35deg 95% 45%)
  color-diverging-2: hsl(40deg 90% 55%)
  color-diverging-3: hsl(45deg 85% 65%)
  color-diverging-4: hsl(50deg 75% 75%)
  color-diverging-5: hsl(55deg 65% 85%)
  color-diverging-6: hsl(165deg 65% 85%)
  color-diverging-7: hsl(170deg 75% 75%)
  color-diverging-8: hsl(175deg 85% 65%)
  color-diverging-9: hsl(180deg 90% 55%)
  color-diverging-10: hsl(185deg 95% 45%)
  color-diverging-11: hsl(190deg 100% 35%)
  
  # Qualitative palette (24 colors) - for categorical data
  color-qualitative-1: hsl(317deg 78% 65%)
  color-qualitative-2: hsl(30deg 100% 50%)
  color-qualitative-3: hsl(240deg 99% 74%)
  color-qualitative-4: hsl(158deg 100% 39%)
  color-qualitative-5: hsl(60deg 100% 38%)
  color-qualitative-6: hsl(18deg 100% 71%)
  color-qualitative-7: hsl(345deg 98% 65%)
  color-qualitative-8: hsl(202deg 100% 43%)
  color-qualitative-9: hsl(170deg 100% 35%)
  color-qualitative-10: hsl(45deg 100% 41%)
  color-qualitative-11: hsl(276deg 77% 66%)
  color-qualitative-12: hsl(87deg 54% 53%)
  color-qualitative-13: hsl(15deg 86% 58%)
  color-qualitative-14: hsl(280deg 65% 55%)
  color-qualitative-15: hsl(195deg 100% 45%)
  color-qualitative-16: hsl(75deg 75% 45%)
  color-qualitative-17: hsl(330deg 85% 60%)
  color-qualitative-18: hsl(50deg 95% 50%)
  color-qualitative-19: hsl(210deg 85% 65%)
  color-qualitative-20: hsl(140deg 70% 45%)
  color-qualitative-21: hsl(25deg 90% 60%)
  color-qualitative-22: hsl(260deg 70% 60%)
  color-qualitative-23: hsl(180deg 80% 40%)
  color-qualitative-24: hsl(100deg 60% 50%)

dark:
  # Primary brand color - lighter for visibility in dark mode
  primary: hsl(180deg 70% 65%)
  
  # Secondary brand color - lighter for visibility in dark mode
  secondary: hsl(45deg 90% 65%)
  
  # Surface/background colors - dark mode
  surface: hsl(180deg 15% 12%)
  background: hsl(180deg 15% 8%)
  
  # Sequential palette - inverted for dark mode (Teal theme)
  color-sequential-1: hsl(180deg 85% 25%)
  color-sequential-2: hsl(180deg 80% 35%)
  color-sequential-3: hsl(180deg 75% 45%)
  color-sequential-4: hsl(180deg 70% 55%)
  color-sequential-5: hsl(180deg 65% 65%)
  color-sequential-6: hsl(180deg 60% 75%)
  color-sequential-7: hsl(180deg 55% 82%)
  color-sequential-8: hsl(180deg 50% 88%)
  color-sequential-9: hsl(180deg 45% 92%)
  
  # Diverging palette - adjusted for dark mode (Amber to Teal)
  color-diverging-1: hsl(190deg 100% 40%)
  color-diverging-2: hsl(185deg 95% 50%)
  color-diverging-3: hsl(180deg 90% 60%)
  color-diverging-4: hsl(175deg 85% 70%)
  color-diverging-5: hsl(170deg 75% 78%)
  color-diverging-6: hsl(55deg 75% 78%)
  color-diverging-7: hsl(50deg 85% 70%)
  color-diverging-8: hsl(45deg 90% 60%)
  color-diverging-9: hsl(40deg 95% 50%)
  color-diverging-10: hsl(35deg 100% 45%)
  color-diverging-11: hsl(30deg 100% 40%)
  
  # Qualitative palette (24 colors) - enhanced for dark mode
  color-qualitative-1: hsl(316deg 79% 62%)
  color-qualitative-2: hsl(27deg 100% 50%)
  color-qualitative-3: hsl(242deg 100% 73%)
  color-qualitative-4: hsl(155deg 100% 38%)
  color-qualitative-5: hsl(59deg 100% 36%)
  color-qualitative-6: hsl(18deg 100% 66%)
  color-qualitative-7: hsl(343deg 98% 61%)
  color-qualitative-8: hsl(203deg 100% 46%)
  color-qualitative-9: hsl(169deg 100% 37%)
  color-qualitative-10: hsl(43deg 100% 41%)
  color-qualitative-11: hsl(277deg 81% 64%)
  color-qualitative-12: hsl(83deg 72% 44%)
  color-qualitative-13: hsl(15deg 90% 60%)
  color-qualitative-14: hsl(280deg 70% 58%)
  color-qualitative-15: hsl(195deg 100% 50%)
  color-qualitative-16: hsl(75deg 80% 48%)
  color-qualitative-17: hsl(330deg 90% 63%)
  color-qualitative-18: hsl(50deg 100% 55%)
  color-qualitative-19: hsl(210deg 90% 68%)
  color-qualitative-20: hsl(140deg 75% 48%)
  color-qualitative-21: hsl(25deg 95% 63%)
  color-qualitative-22: hsl(260deg 75% 63%)
  color-qualitative-23: hsl(180deg 85% 45%)
  color-qualitative-24: hsl(100deg 65% 53%)
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
