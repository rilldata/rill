import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import {
  ResourceKind,
  UserFacingResourceKinds,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function handleEntityCreate(kind: ResourceKind) {
  if (!(kind in ResourceKindMap)) return;

  // Get the path for the new file
  const allNames =
    kind === ResourceKind.Source || kind === ResourceKind.Model
      ? // sources and models share the name
        [
          ...fileArtifacts.getNamesForKind(ResourceKind.Source),
          ...fileArtifacts.getNamesForKind(ResourceKind.Model),
        ]
      : fileArtifacts.getNamesForKind(kind);
  const { folderName, baseFileName, extension, content } =
    ResourceKindMap[kind];
  const newName = getName(baseFileName, allNames);
  const newPath = `${folderName}/${newName}${extension}`;

  // Create the new file
  const instanceId = get(runtime).instanceId;
  await runtimeServicePutFile(instanceId, {
    path: newPath,
    blob: content,
    create: true,
    createOnly: true,
  });

  // Return the path to the new file, so we can navigate the user to it
  return `/files/${newPath}`;
}

const ResourceKindMap: Record<
  UserFacingResourceKinds,
  {
    folderName: string;
    baseFileName: string;
    extension: string;
    content: string;
  }
> = {
  [ResourceKind.Source]: {
    folderName: "sources",
    baseFileName: "source",
    extension: ".yaml",
    content: "", // This is constructed in the `features/sources/modal` directory
  },
  [ResourceKind.Connector]: {
    folderName: "connectors",
    baseFileName: "connector",
    extension: ".yaml",
    content: "", // This is constructed in the `features/connectors` directory
  },
  [ResourceKind.Model]: {
    folderName: "models",
    baseFileName: "model",
    extension: ".sql",
    content: `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

SELECT 'Hello, World!' AS Greeting`,
  },
  [ResourceKind.MetricsView]: {
    folderName: "metrics",
    baseFileName: "metrics_view",
    extension: ".yaml",
    content: `# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: example_table # Choose a table to underpin your metrics
timeseries: timestamp_column # Choose a timestamp column (if any) from your table

dimensions:
  - column: category
    label: "Category"
    description: "Description of the dimension"

measures:
  - expression: "SUM(revenue)"
    label: "Total Revenue"
    description: "Total revenue generated"
`,
  },
  [ResourceKind.Explore]: {
    folderName: "explore-dashboards",
    baseFileName: "explore",
    extension: ".yaml",
    content: `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explores

type: explore

title: "My metrics dashboard"
metrics_view: example_metrics_view # Choose a metrics view to underpin the dashboard

dimensions: '*'
measures: '*'
`,
  },
  [ResourceKind.API]: {
    folderName: "apis",
    baseFileName: "api",
    extension: ".yaml",
    content: `# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

type: api

metrics_sql: |
  select measure, dimension from metrics_view
`,
  },
  [ResourceKind.Component]: {
    folderName: "components",
    baseFileName: "component",
    extension: ".yaml",
    content: `# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components
    
type: component

data:
  sql: |
    SELECT * FROM (VALUES 
      ('Monday', 300),
      ('Tuesday', 150),
      ('Wednesday', 200),
      ('Thursday', 400),
      ('Friday', 650),
      ('Saturday', 575),
      ('Sunday', 500)
    ) AS t(day_of_week, revenue)

vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": { "name": "table" },
    "mark": "line",
    "width": "container",
    "encoding": {
      "x": {
        "field": "day_of_week",
        "type": "ordinal",
        "axis": { "title": "Day of the Week" },
        "sort": [
          "Monday",
          "Tuesday",
          "Wednesday",
          "Thursday",
          "Friday",
          "Saturday",
          "Sunday"
        ]
      },
      "y": {
        "field": "revenue",
        "type": "quantitative",
        "axis": { "title": "Revenue" }
      }
    }
  }`,
  },
  [ResourceKind.Canvas]: {
    folderName: "canvas-dashboards",
    baseFileName: "canvas",
    extension: ".yaml",
    content: `type: canvas
title: "Canvas Dashboard"
columns: 24
gap: 2

items:
  - component:
      markdown:
        content: "First Component"
        css:
          font-size: "40px"
          background-color: "#fff"
    width: 4
    height: 3
    x: 2
    y: 1
`,
  },
  [ResourceKind.Theme]: {
    folderName: "themes",
    baseFileName: "theme",
    extension: ".yaml",
    content: `# Theme YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/themes

type: theme

colors:
  primary: plum
  secondary: violet 
`,
  },
  [ResourceKind.Report]: {
    folderName: "reports",
    baseFileName: "report",
    extension: ".yaml",
    content: `# Report YAML
# Reference documentation: TODO

type: report

...
`,
  },
  [ResourceKind.Alert]: {
    folderName: "alerts",
    baseFileName: "alert",
    extension: ".yaml",
    content: `# Alert YAML
# Reference documentation: TODO

type: alert

...
`,
  },
};
