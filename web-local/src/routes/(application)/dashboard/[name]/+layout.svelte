<script lang="ts">
  import { page } from "$app/stores";
  import ModelInspector from "@rilldata/web-common/features/models/workspace/inspector/ModelInspector.svelte";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    Inspector,
    WorkspaceContainer,
  } from "@rilldata/web-local/lib/components/workspace";
  import DashboardWorkspaceHeader from "@rilldata/web-local/lib/components/workspace/explore/workspace-header/DashboardWorkspaceHeader.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { parseDocument } from "yaml";
  export let data;
  /** default to saying this is a valid dashboard unless we return something else. */
  $: validDashboard =
    data?.validDashboard === undefined ? true : data.validDashboard;

  $: configName = data.configName;
  $: entry = data.entry;

  $: thisConfigFile = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    getFilePathFromNameAndType(configName, EntityType.MetricsDefinition)
  );

  $: parsedConfigFile = parseDocument(
    $thisConfigFile?.data?.blob || "{}"
  ).toJS();

  // if the parsedConfig file has a model defined, then we need to show the model inspector.
  $: modelDefined = parsedConfigFile?.model?.length > 0;

  $: thisModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    parsedConfigFile?.model
  );

  $: metricViewName = $page.params.name;
  $: displayName = entry?.metricsView?.label || entry?.metricsView?.name;
  $: view =
    $page?.url?.pathname === `/dashboard/${metricViewName}`
      ? "dashboard"
      : $page?.url?.pathname === `/dashboard/${metricViewName}/edit`
      ? "config"
      : "model";

  // need to only render if the referenced model is loaded,
  // or if no model is defined for this dashboard.
</script>

<WorkspaceContainer
  assetID={metricViewName}
  bgClass="bg-white"
  viewHasInspector={true}
  inspector={view !== "dashboard"}
>
  <DashboardWorkspaceHeader
    slot="header"
    {displayName}
    {metricViewName}
    showModelToggle={modelDefined}
    showDashboardToggle={validDashboard}
  />
  <slot />
  {#if view !== "dashboard"}
    <Inspector>
      <ModelInspector modelName={parsedConfigFile?.model} />
    </Inspector>
  {/if}
</WorkspaceContainer>
