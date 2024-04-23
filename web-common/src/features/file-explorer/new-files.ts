import { fetchAllFileNames } from "@rilldata/web-common/features/entity-management/file-selectors";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import {
  ResourceKind,
  UserFacingResourceKinds,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function handleEntityCreate(kind: ResourceKind) {
  if (!(kind in ResourceKindMap)) return;
  const instanceId = get(runtime).instanceId;
  const allNames = await fetchAllFileNames(queryClient, instanceId, false);
  const { name, extension, baseContent } = ResourceKindMap[kind];
  const newName = getName(name, allNames);
  const newPath = `${name + "s"}/${newName}${extension}`;

  await runtimeServicePutFile(instanceId, newPath, {
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
    baseContent: "",
  },
  [ResourceKind.Model]: {
    name: "model",
    extension: ".sql",
    baseContent: `SELECT 'Hello, World!' AS Greeting

-- The \`@kind: model\` decorator registers your Model if this file is moved out of the \`/models\` directory.
--@kind: model`,
  },
  [ResourceKind.MetricsView]: {
    name: "dashboard",
    extension: ".yaml",
    baseContent: `# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

table: example_table # Choose a table to underpin your dashboard

title: "Dashboard Title"

timeseries: timestamp # Replace with an actual timestamp column (if any) from your table

# Configure the dashboard's dimensions...
dimensions:
  - column: category
    label: "Category"
    description: "Description of the dimension"

# Configure the dashboard's measures...
measures:
  - expression: "SUM(revenue)"
    label: "Total Revenue"
    description: "Total revenue generated"

# \`kind: metrics_view\` registers your Dashboard if this file is moved out of the \`/dashboards\` directory.
kind: metrics_view
`,
  },
  [ResourceKind.API]: {
    name: "api",
    extension: ".yaml",
    baseContent: `kind: api

sql:
  select ...
`,
  },
  [ResourceKind.Chart]: {
    name: "chart",
    extension: ".yaml",
    baseContent: `kind: chart
data:
  metrics_sql: |
    SELECT advertiser_name, AGGREGATE(measure_2)
    FROM Bids_Sample_Dash
    GROUP BY advertiser_name
    ORDER BY measure_2 DESC
    LIMIT 20

vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": {"name": "table"},
    "mark": "bar",
    "width": "container",
    "encoding": {
      "x": {"field": "advertiser_name", "type": "nominal"},
      "y": {"field": "measure_2", "type": "quantitative"}
    }
  }`,
  },
  [ResourceKind.Dashboard]: {
    name: "custom-dashboard",
    extension: ".yaml",
    baseContent: `kind: dashboard
columns: 10
gap: 2`,
  },
  [ResourceKind.Theme]: {
    name: "theme",
    extension: ".yaml",
    baseContent: `kind: theme
colors:
  primary: crimson 
  secondary: lime 
`,
  },
  [ResourceKind.Report]: {
    name: "report",
    extension: ".yaml",
    baseContent: `kind: report

...
`,
  },
  [ResourceKind.Alert]: {
    name: "alert",
    extension: ".yaml",
    baseContent: `kind: alert

...
`,
  },
};
