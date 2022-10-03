<script lang="ts">
  import Model from "@rilldata/web-local/lib/components/workspace/Model.svelte";

  export let data;
</script>

<svelte:head>
  <!-- TODO: add the model name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<Model modelId={data.modelId} />
