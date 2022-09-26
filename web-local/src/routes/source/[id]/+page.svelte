<script lang="ts">
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$web-local/lib/application-state-stores/application-store";
  import SourceWorkspace from "$web-local/lib/components/workspace/source/SourceWorkspace.svelte";

  export let data;

  $: sourceId = data.sourceId;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.Table,
    sourceId,
  ]);
</script>

<svelte:head>
  <!-- TODO: add the source name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<SourceWorkspace />
