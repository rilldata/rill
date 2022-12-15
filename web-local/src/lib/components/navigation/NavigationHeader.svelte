<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../application-config";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import ContextButton from "../column-profile/ContextButton.svelte";
  import AddIcon from "../icons/Add.svelte";
  const dispatch = createEventDispatcher();

  export let show = true;
  export let tooltipText: string;
  /** The CSS ID used for tests for the context button */
  export let contextButtonID: string = undefined;
</script>

<div
  class="pl-4 pb-3 pr-3 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide|local={{ duration: LIST_SLIDE_DURATION }}
>
  <CollapsibleSectionTitle tooltipText={"models"} bind:active={show}>
    <div class="flex flex-row items-center gap-x-2">
      <slot />
    </div>
  </CollapsibleSectionTitle>
  <ContextButton
    id={contextButtonID}
    {tooltipText}
    on:click={() => {
      dispatch("add");
    }}
    width={24}
    height={24}
    rounded
  >
    <AddIcon />
  </ContextButton>
</div>
