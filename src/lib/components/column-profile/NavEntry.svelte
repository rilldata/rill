<script lang="ts">
  // FIXME: we should probably rename this to AssetNavigationElement.svelte or something like that.
  import { createEventDispatcher } from "svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import ExpanderButton from "$lib/components/column-profile/ExpanderButton.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  const { shiftClickAction } = createShiftClickAction();
  const dispatch = createEventDispatcher();

  export let entityType: EntityType;
  export let expanded = true;
  export let selected = false;
  export let hovered = false;
</script>

<div
  on:mouseenter={() => {
    hovered = true;
  }}
  on:mouseleave={() => {
    hovered = false;
  }}
  style:height="24px"
  style:grid-template-columns="[left-control] max-content [body] auto
  [contextual-information] max-content"
  class="
        {selected ? 'bg-gray-100' : 'bg-transparent'}
        grid
        grid-flow-col
        gap-2
        items-center
        hover:bg-gray-200
        pl-4 pr-4 
    "
>
  <ExpanderButton
    rotated={expanded}
    on:click={() => {
      dispatch("expand");
    }}
  >
    <CaretDownIcon size="14px" />
  </ExpanderButton>
  <Tooltip location="right">
    <button
      use:shiftClickAction
      on:shift-click
      on:click={(evt) => {
        dispatch("select");
        if (
          entityType == EntityType.Table ||
          (entityType == EntityType.Model && selected)
        ) {
          dispatch("expand");
        }
      }}
      on:focus={() => {
        hovered = true;
      }}
      on:blur={() => {
        hovered = false;
      }}
      style:grid-column="body"
      style:grid-template-columns="[icon] max-content [text] 1fr"
      class="
                w-full 
                justify-start
                text-left 
                grid 
                items-center
                p-0"
    >
      <div
        style:grid-column="text"
        class="
                    w-full
                    justify-self-auto
                    text-ellipsis 
                    overflow-hidden 
                    whitespace-nowrap"
      >
        <slot />
      </div>
    </button>
    <TooltipContent slot="tooltip-content">
      <slot name="tooltip-content" />
    </TooltipContent>
  </Tooltip>
  <div style:grid-column="contextual-information" class="justify-self-end">
    <slot name="contextual-information" />
  </div>
</div>
