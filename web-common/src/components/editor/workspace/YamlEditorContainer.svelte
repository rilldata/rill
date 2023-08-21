<!-- @component 
The full size container is meant to be embedded in a Workspace.svelte component.
It will show an error message if passed in.
-->

<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slide } from "svelte/transition";

  export let errorMessage = "";
  export let height = "calc(100vh - var(--header-height))";
</script>

<div class="flex flex-col w-full h-full content-stretch" style:height>
  <div class="grow bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={errorMessage}
      class:border-red-500={errorMessage}
    >
      <slot />
    </div>
  </div>
  {#if errorMessage}
    <div
      role="status"
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
      class="editor-error ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{errorMessage}
      </div>
    </div>
  {/if}
</div>
