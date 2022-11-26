<script>
  import { createShiftClickAction } from "@rilldata/web-local/lib/util/shift-click-action";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../application-config";

  const dispatch = createEventDispatcher();
  const { shiftClickAction } = createShiftClickAction();

  export let active = false;
  export let emphasize = false;
  export let hideRight = false;
</script>

<div>
  <button
    class="
        select-none	
        flex 
        space-between 
        gap-2
        hover:bg-gray-100 
        focus:bg-gray-100
        focus:ring-gray-500
        focus:outline-gray-300 flex-1
        justify-between w-full"
    class:bg-gray-50={active}
    use:shiftClickAction
    on:shift-click
    on:click={() => {
      dispatch("select");
    }}
  >
    <div class="flex gap-2 items-baseline" style:min-width="0px">
      <div class="self-center flex items-center ui-copy-icon-muted">
        <slot name="icon" />
      </div>
      <div class:font-bold={emphasize} class="text-left" style:min-width="0px">
        <slot name="left" />
      </div>
    </div>
    <div class:hidden={hideRight} class="flex gap-x-2 items-center">
      <slot name="right" />
      <slot name="context-button" />
    </div>
  </button>
  {#if active && $$slots["details"]}
    <div
      class="w-full"
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
    >
      <slot name="details" />
    </div>
  {/if}
</div>
