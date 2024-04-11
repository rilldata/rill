<!-- @component 
The full size container is meant to be embedded in a Workspace.svelte component.
It will show an error message if passed in.
-->

<script lang="ts">
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";

  export let error: LineStatus | V1ReconcileError | undefined = undefined;
</script>

<div class="flex flex-col size-full bg-white">
  <div
    class="overflow-auto size-full border-white"
    class:border-b-hidden={error}
    class:border-red-500={error}
  >
    <slot />
  </div>

  {#if error}
    <div
      role="status"
      transition:slide={{ duration: LIST_SLIDE_DURATION }}
      class="m-2 editor-error ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-2 max-h-72"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{error.message}
      </div>
    </div>
  {/if}
</div>
