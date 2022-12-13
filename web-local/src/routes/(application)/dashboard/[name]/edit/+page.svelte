<script lang="ts">
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";
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

{#if yaml}
  <MetricsDefinitionWorkspace {metricsDefName} {nonStandardError} {yaml} />
{/if}
