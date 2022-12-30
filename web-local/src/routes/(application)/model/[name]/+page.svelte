<script lang="ts">
  import { ModelWorkspace } from "@rilldata/web-local/lib/components/workspace";

  export let data;
</script>

<svelte:head>
  <title>Rill Developer | {data.modelName}</title>
</svelte:head>

<ModelWorkspace modelName={data.modelName} />
