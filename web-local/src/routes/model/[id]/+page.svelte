<script lang="ts">
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$web-local/lib/application-state-stores/application-store";
  import Model from "$web-local/lib/components/workspace/Model.svelte";
  import ModelWorkspaceHeader from "$web-local/lib/components/workspace/ModelWorkspaceHeader.svelte";

  export let data;

  $: modelId = data.modelId;

  $: dataModelerService.dispatch("setActiveAsset", [EntityType.Model, modelId]);
</script>

<svelte:head>
  <!-- TODO: add the model name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<ModelWorkspaceHeader />
<Model />
