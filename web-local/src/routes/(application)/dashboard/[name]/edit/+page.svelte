<script lang="ts">
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";
  import WorkspaceBody from "@rilldata/web-local/lib/components/workspace/core/WorkspaceBody.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { parseDocument } from "yaml";

  export let data;

  $: metricsDefName = data.metricsDefName;
  $: instanceId = $runtimeStore.instanceId;

  $: dashboardYAML = useRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );

  $: yaml = $dashboardYAML.data?.blob || "";
  $: nonStandardError = data.error;
  $: parsedYAML = parseDocument($dashboardYAML?.data?.blob || "{}").toJS();
  $: console.log(parsedYAML);
  $: hasModel = parsedYAML?.model?.length > 0;
</script>

<svelte:head>
  <title>Rill Developer | {metricsDefName}</title>
</svelte:head>

<WorkspaceBody right={hasModel}>
  {#if yaml}
    <!-- FIXME: change this to be DashboardConfigBody.svelte -->
    <MetricsDefinitionWorkspace {metricsDefName} {nonStandardError} {yaml} />
  {/if}
</WorkspaceBody>
