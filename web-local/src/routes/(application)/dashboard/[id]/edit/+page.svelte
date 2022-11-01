<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { Dialog } from "@rilldata/web-local/lib/components/modal";

  export let data;

  $: metricsDefId = data.metricsDefId;
  $: error = data.error;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsDefinition,
    metricsDefId,
  ]);

  function onCancel() {
    error = "";
  }
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<MetricsDefinitionWorkspace metricsDefId={data.metricsDefId} />

{#if error}
  <Dialog compact on:cancel={onCancel}>
    <svelte:fragment slot="title">Unable to display Dashboard</svelte:fragment>
    <svelte:fragment slot="body">{error}</svelte:fragment>
    <Button on:click={onCancel} slot="footer" type="text">Close</Button>
  </Dialog>
{/if}
