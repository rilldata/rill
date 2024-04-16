import { fetchAllFileNames } from "@rilldata/web-common/features/entity-management/file-selectors";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function handleEntityCreate(kind: ResourceKind) {
  if (!(kind in ResourceKindMap)) return;
  const instanceId = get(runtime).instanceId;
  const allNames = await fetchAllFileNames(queryClient, instanceId);
  const { name, folder, baseContent, extension } = ResourceKindMap[kind];
  const newName = getName(name, allNames);

  const newPath = `${folder ?? name + "s"}/${newName}${extension ?? ".yaml"}`;

  await runtimeServicePutFile(instanceId, newPath, {
    blob: baseContent,
    create: true,
    createOnly: true,
  });
  return `/files//${newPath}`;
}

const ResourceKindMap: Record<
  ResourceKind,
  {
    name: string;
    folder?: string; // adds "s" to name by default
    baseContent: string;
    extension?: string;
  }
> = {
  [ResourceKind.ProjectParser]: { baseContent: "", name: "" },
  [ResourceKind.Source]: {
    name: "source",
    baseContent: "",
  },
  [ResourceKind.Model]: {
    name: "model",
    extension: ".sql",
    baseContent: `-- @kind: model
select ...
`,
  },
  [ResourceKind.MetricsView]: {
    name: "dashboard",
    baseContent: `kind: metrics_view

`,
  },
  [ResourceKind.API]: {
    name: "api",
    baseContent: `kind: api

sql:
  select ...
`,
  },
  [ResourceKind.Chart]: {
    name: "chart",
    baseContent: `kind: chart

...
`,
  },
  [ResourceKind.Dashboard]: {
    name: "custom-dashboard",
    baseContent: `kind: dashboard
    
...`,
  },
  [ResourceKind.Theme]: {
    name: "theme",
    baseContent: `kind: theme
colors:
  primary: crimson 
  secondary: lime 
`,
  },
  [ResourceKind.Report]: {
    name: "report",
    baseContent: `kind: report

...
`,
  },
  [ResourceKind.Alert]: {
    name: "alert",
    baseContent: `kind: alert

...
`,
  },
};
