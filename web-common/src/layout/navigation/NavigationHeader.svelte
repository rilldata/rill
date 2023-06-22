<script lang="ts">
  import { slide } from "svelte/transition";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "../config";

  export let show = true;
  export let toggleText: string;
</script>

<div
  class="pl-4 pb-1 pr-2 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide|local={{ duration: LIST_SLIDE_DURATION }}
>
  <CollapsibleSectionTitle tooltipText={toggleText} bind:active={show}>
    <div class="flex flex-row items-center gap-x-2">
      <slot />
    </div>
  </CollapsibleSectionTitle>
</div>
