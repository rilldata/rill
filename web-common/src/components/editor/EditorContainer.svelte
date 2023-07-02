<!-- @component 
The full size container is meant to be embedded in a Workspace.svelte component.
It will show an error message if passed in.
-->

<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";
  import CancelCircle from "../icons/CancelCircle.svelte";
  import type { LineStatus } from "./line-status/state";

  export let error: LineStatus | V1ReconcileError = undefined;
  export let hasContent = false;
  export let height = "calc(100vh - var(--header-height))";
</script>

<div class="flex flex-col w-full h-full content-stretch" style:height>
  <div class="grow flex bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={error && hasContent}
      class:border-red-500={error && hasContent}
    >
      <slot />
    </div>
  </div>
  {#if error && hasContent}
    <div
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
      class="ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{error.message}
      </div>
    </div>
  {/if}
</div>
