<script lang="ts">
  import AddIcon from "@rilldata/web-common/components/icons/Add.svelte";
  import ContextButton from "@rilldata/web-local/lib/components/column-profile/ContextButton.svelte";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "../config";

  export let show = true;
  export let tooltipText: string;
  export let toggleText = "models";
  /** The CSS ID used for tests for the context button */
  export let contextButtonID: string = undefined;
  export let canAddAsset = true;

  const dispatch = createEventDispatcher();
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
  {#if canAddAsset}
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
  {/if}
</div>
