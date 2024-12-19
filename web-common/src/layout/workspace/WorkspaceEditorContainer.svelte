<script lang="ts">
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slide } from "svelte/transition";

  export let error: LineStatus | undefined = undefined;
  export let showError = true;
</script>

<div class="flex flex-col size-full gap-y-1">
  <div
    class="size-full border overflow-y-hidden rounded-[2px] bg-background flex flex-col items-center justify-center"
    class:!border-red-500={error}
  >
    <slot />
  </div>

  {#if error && showError}
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
