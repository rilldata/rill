<script lang="ts">
  import { slide } from "svelte/transition";
  import CancelCircle from "../../components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "../../layout/config";

  export let errorMessage: string;
</script>

<div
  role="status"
  transition:slide={{ duration: LIST_SLIDE_DURATION }}
  class="editor-error ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5"
>
  <div class="flex gap-x-2 items-center">
    <CancelCircle />{errorMessage}
  </div>
</div>
