<script lang="ts">
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$web-local/lib/application-state-stores/application-store";
  import MetricsDefWorkspace from "$web-local/lib/components/workspace/metrics-def/MetricsDefWorkspace.svelte";
  import MetricsDefWorkspaceHeader from "$web-local/lib/components/workspace/metrics-def/MetricsDefWorkspaceHeader.svelte";

  export let data;

  $: metricsDefId = data.metricsDefId;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsDefinition,
    metricsDefId,
  ]);
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<MetricsDefWorkspaceHeader metricsDefId={data.metricsDefId} />
<MetricsDefWorkspace metricsDefId={data.metricsDefId} />
