<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import Explore from "@rilldata/web-local/lib/components/workspace/explore/Explore.svelte";
  export let data;

  $: metricsDefId = data.metricsDefId;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsExplorer,
    metricsDefId,
  ]);
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<!-- 
  temporary workaround to clear out non-white bg space. 
  once the q4 application refactor is complete let's remove this.
-->
<!-- <div
  class="bg-white absolute left-0 top-0 right-0 bottom-0 w-screen h-screen"
  style:border="4px solid black"
/> -->
<Explore {metricsDefId} />
