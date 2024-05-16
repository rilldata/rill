<!-- @component 
The full size container is meant to be embedded in a Workspace.svelte component.
It will show an error message if passed in.
-->

<script lang="ts">
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slide } from "svelte/transition";

  export let error: LineStatus | undefined = undefined;
</script>

<div class="flex flex-col w-full h-full">
  <div
    class="size-full overflow-auto border border-white bg-white"
    class:!border-red-500={error}
    class:border-b-0={error}
  >
    <slot />
  </div>

  {#if error}
    <div
      role="status"
      transition:slide={{ duration: LIST_SLIDE_DURATION }}
      class="editor-error ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5 max-h-72 overflow-auto"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{error.message}
      </div>
    </div>
  {/if}
</div>
