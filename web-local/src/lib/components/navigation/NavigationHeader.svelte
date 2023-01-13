<script lang="ts">
  import AddIcon from "@rilldata/web-common/components/icons/Add.svelte";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../application-config";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import ContextButton from "../column-profile/ContextButton.svelte";

  export let show = true;
  export let tooltipText: string;
  export let toggleText = "models";
  export let showContextButton = true;
  /** The CSS ID used for tests for the context button */
  export let contextButtonID: string = undefined;

  const dispatch = createEventDispatcher();
</script>

<div
  style:height="28px"
  class="pl-4 pb-1 pr-3 grid justify-between"
  style="grid-template-columns: auto max-content;"
  out:slide|local={{ duration: LIST_SLIDE_DURATION }}
>
  <CollapsibleSectionTitle tooltipText={toggleText} bind:active={show}>
    <div class="flex flex-row items-center gap-x-2">
      <slot />
    </div>
  </CollapsibleSectionTitle>
  {#if showContextButton}
    <slot name="context-button">
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
    </slot>
  {/if}
</div>
