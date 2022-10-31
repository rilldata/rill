<script lang="ts">
  import ModelWorkspace from "@rilldata/web-local/lib/components/workspace/model/ModelWorkspace.svelte";

  export let data;
</script>

<svelte:head>
  <!-- TODO: add the model name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<ModelWorkspace modelID={data.modelID} />
