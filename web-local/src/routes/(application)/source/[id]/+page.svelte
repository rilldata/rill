<script lang="ts">
  import SourceWorkspace from "@rilldata/web-local/lib/components/workspace/source/SourceWorkspace.svelte";

  export let data;
</script>

<svelte:head>
  <!-- TODO: add the source name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<SourceWorkspace sourceId={data.sourceId} />
