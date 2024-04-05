<script lang="ts">
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { page } from "$app/stores";
  import HorizontalSplitter from "@rilldata/web-common/layout/workspace/HorizontalSplitter.svelte";

  export let fade = false;

  $: context = $page.url.pathname;
  $: workspaceLayout = workspaces.get(context);
  $: tableHeight = workspaceLayout.table.height;
</script>

<div
  class="p-5 w-full relative flex flex-none flex-col gap-2"
  style:height="{$tableHeight}px"
  style:max-height="75%"
>
  <Resizer max={600} direction="NS" side="top" bind:dimension={$tableHeight}>
    <HorizontalSplitter />
  </Resizer>

  <div class="table-wrapper" class:brightness-90={fade}>
    <slot />
  </div>

  <slot name="error" />
</div>

<style lang="postcss">
  .table-wrapper {
    transition: filter 200ms;
    @apply relative rounded w-full overflow-hidden border-gray-200 border-2 h-full;
  }
</style>
