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
  export let height = "calc(100vh - var(--header-height))";
</script>

<div class="flex flex-col w-1/2 h-full content-stretch" style:height>
  <div class="grow bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={error}
      class:border-red-500={error}
    >
      <slot />
    </div>
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
