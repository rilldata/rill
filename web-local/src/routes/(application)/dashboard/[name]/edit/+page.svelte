<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";

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
