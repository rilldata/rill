<script lang="ts">
  import { SourceWorkspace } from "@rilldata/web-local/lib/components/workspace";

  export let data;
</script>

<svelte:head>
  <title>Rill Developer | {data.sourceName}</title>
</svelte:head>

<SourceWorkspace sourceName={data.sourceName} />
