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
  const instanceId = get(runtime).instanceId;
  const allNames =
    kind === ResourceKind.Source || kind === ResourceKind.Model
      ? // sources and models share the name
      [
        ...fileArtifacts.getNamesForKind(ResourceKind.Source),
        ...fileArtifacts.getNamesForKind(ResourceKind.Model),
      ]
      : fileArtifacts.getNamesForKind(kind);
  const { name, extension, baseContent } = ResourceKindMap[kind];
  const newName = getName(name, allNames);
  const newPath = `${name + "s"}/${newName}${extension}`;

  await runtimeServicePutFile(instanceId, {
    path: newPath,
    blob: baseContent,
    create: true,
    createOnly: true,
  });
  return `/files/${newPath}`;
}

const ResourceKindMap: Record<
  UserFacingResourceKinds,
  {
    name: string;
    extension: string;
    baseContent: string;
  }
> = {
  [ResourceKind.Source]: {
    name: "source",
    extension: ".yaml",
    baseContent: "", // This is constructed in the `features/sources/modal` directory
  },
  [ResourceKind.Model]: {
    name: "model",
    extension: ".sql",
    baseContent: `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

SELECT 'Hello, World!' AS Greeting`,
  },
  [ResourceKind.MetricsView]: {
    name: "dashboard",
    extension: ".yaml",
    baseContent: `# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

type: metrics_view

title: "Dashboard Title"
table: example_table # Choose a table to underpin your dashboard
timeseries: timestamp_column # Select an actual timestamp column (if any) from your table

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
  [ResourceKind.API]: {
    name: "api",
    extension: ".yaml",
    baseContent: `# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

type: api

metrics_sql: |
  select measure, dimension from metrics_view
`,
  },
  [ResourceKind.Component]: {
    name: "chart",
    extension: ".yaml",
    baseContent: `# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts
    
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
  [ResourceKind.Dashboard]: {
    name: "custom-dashboard",
    extension: ".yaml",
    baseContent: `type: dashboard
columns: 10
gap: 2

items:
  - component:
      type: markdown
      markdown: First component
    width: 4
    height: 3
    x: 2
    y: 1
`,
  },
  [ResourceKind.Theme]: {
    name: "theme",
    extension: ".yaml",
    baseContent: `# Theme YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/themes

type: theme

colors:
  primary: plum
  secondary: violet 
`,
  },
  [ResourceKind.Report]: {
    name: "report",
    extension: ".yaml",
    baseContent: `# Report YAML
# Reference documentation: TODO

type: report

...
`,
  },
  [ResourceKind.Alert]: {
    name: "alert",
    extension: ".yaml",
    baseContent: `# Alert YAML
# Reference documentation: TODO

type: alert

...
`,
  },
};
