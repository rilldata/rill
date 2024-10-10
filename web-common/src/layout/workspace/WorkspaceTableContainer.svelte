<script lang="ts">
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { slide } from "svelte/transition";

  export let fade = false;
  export let filePath: string;

  $: workspaceLayout = workspaces.get(filePath);
  $: tableHeight = workspaceLayout.table.height;
</script>

<div
  transition:slide
  class="w-full relative flex flex-none flex-col"
  style:height="{$tableHeight}px"
  style:max-height="75%"
>
  <Resizer
    absolute={false}
    max={600}
    direction="NS"
    side="top"
    bind:dimension={$tableHeight}
  />
  <div class="table-wrapper" class:brightness-90={fade}>
    <slot />
  </div>

  <slot name="error" />
</div>

<style lang="postcss">
  .table-wrapper {
    transition: filter 200ms;
    @apply relative rounded-[2px] w-full overflow-hidden border h-full;
  }
</style>
