<script lang="ts">
  import { EntityType } from "../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "../../../lib/application-state-stores/application-store";
  import Model from "../../../lib/components/workspace/Model.svelte";
  import ModelWorkspaceHeader from "../../../lib/components/workspace/ModelWorkspaceHeader.svelte";

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
