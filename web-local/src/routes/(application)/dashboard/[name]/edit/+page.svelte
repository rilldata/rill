<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";

  export let data;

  $: metricsDefId = data.metricsDefId;
  $: nonStandardError = data.error;
  $: yaml = data.yaml;
  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsDefinition,
    metricsDefId,
  ]);
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<MetricsDefinitionWorkspace
  metricsDefId={data.metricsDefId}
  {nonStandardError}
  {yaml}
/>
