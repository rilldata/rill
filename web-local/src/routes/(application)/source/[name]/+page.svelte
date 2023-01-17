<script lang="ts">
  import { SourceWorkspace } from "@rilldata/web-common/features/sources";

  export let data;
</script>

<svelte:head>
  <title>Rill Developer | {data.sourceName}</title>
</svelte:head>

<SourceWorkspace
  sourceName={data.sourceName}
  embedded={data?.embedded}
  path={data?.path}
/>
