<script lang="ts">
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";
  import WorkspaceBody from "@rilldata/web-local/lib/components/workspace/core/WorkspaceBody.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";

  export let data;

  $: metricsDefName = data.metricsDefName;
  $: instanceId = $runtimeStore.instanceId;

  $: dashboardYAML = useRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );

  $: yaml = $dashboardYAML.data?.blob || "";
  $: nonStandardError = data.error;
</script>

<svelte:head>
  <title>Rill Developer | {metricsDefName}</title>
</svelte:head>

<WorkspaceBody>
  {#if yaml}
    <MetricsDefinitionWorkspace {metricsDefName} {nonStandardError} {yaml} />
  {/if}
</WorkspaceBody>
