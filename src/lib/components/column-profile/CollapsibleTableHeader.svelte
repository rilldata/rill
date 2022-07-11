<script lang="ts">
  import { tick } from "svelte/internal";
  import { tweened } from "svelte/motion";
  import { cubicInOut as easing } from "svelte/easing";
  import { format } from "d3-format";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import ExpanderButton from "$lib/components/column-profile/ExpanderButton.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { guidGenerator } from "$lib/util/guid";
  import notificationStore from "$lib/components/notifications/";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { createEventDispatcher } from "svelte";
  import { createCommandClickAction } from "$lib/util/command-click-action";
  import { createShiftClickAction } from "$lib/util/shift-click-action";

  export let entityType: EntityType;
  export let name: string;
  export let cardinality: number;
  export let showRows = true;
  export let sizeInBytes: number = undefined;
  export let active = false;
  export let menuX: number = undefined;
  export let menuY: number = undefined;
  export let show = false;
  export let contextMenuOpen = false;
  export let contextMenu: any;

  const dispatch = createEventDispatcher();
  const { commandClickAction } = createCommandClickAction();
  const { shiftClickAction } = createShiftClickAction();

  const formatInteger = format(",");

  let cardinalityTween = tweened(cardinality, { duration: 600, easing });
  let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });

  $: cardinalityTween.set(cardinality || 0);
  $: interimCardinality = ~~$cardinalityTween;
  $: sizeTween.set(sizeInBytes || 0);

  let selectingColumns = false;
  let selectedColumns = [];

  const contextButtonId = guidGenerator();

  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  let hovered = false;
  $: showEntityDetails = hovered || active || contextMenuOpen;

  const commandClickHandler = () => {
    if (entityType == EntityType.Table) {
      dispatch("query");
    }
  };

  const shiftClickHandler = async () => {
    await navigator.clipboard.writeText(name);
    notificationStore.send({ message: `copied "${name}" to clipboard` });
  };

  const clickEntityNameHandler = () => {
    dispatch("select");
    if (
      entityType == EntityType.Table ||
      (entityType == EntityType.Model && active)
    ) {
      show = !show;
    }
  };

  const clickContextButtonHandler = async (event) => {
    contextMenuOpen = !contextMenuOpen;
    menuX = event.clientX;
    menuY = event.clientY;

    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        contextMenuOpen = false;
      }, contextMenu);
    }
  };
</script>

<Tooltip location="right">
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
    class=" grid grid-flow-col gap-2 items-center hover:bg-gray-200 pl-4 pr-4 {active
      ? 'bg-gray-100'
      : 'bg-transparent'}
    "
  >
    <ExpanderButton rotated={show} on:click={() => dispatch("expand")}>
      <CaretDownIcon size="14px" />
    </ExpanderButton>
    <button
      use:commandClickAction
      on:command-click={commandClickHandler}
      use:shiftClickAction
      on:shift-click={shiftClickHandler}
      on:click={clickEntityNameHandler}
      on:focus={() => (hovered = true)}
      on:blur={() => (hovered = false)}
      style:grid-column="body"
      style:grid-template-columns="[icon] max-content [text] 1fr"
      class="w-full justify-start text-left grid items-center p-0"
    >
      <div
        style:grid-column="text"
        class="w-full justify-self-auto text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <!-- note: the classes in this span are also used for UI tests. -->
        <span
          class="collapsible-table-summary-title w-full"
          class:is-active={active}
          class:font-bold={active}
          class:italic={selectingColumns}
        >
          {#if name.split(".").length > 1}
            {name.split(".").slice(0, -1).join(".")}<span
              class="text-gray-500 italic pl-1"
              >.{name.split(".").slice(-1).join(".")}</span
            >
          {:else}
            {name}
          {/if}
          {#if selectingColumns}&nbsp;<span class="font-bold"> *</span>{/if}
        </span>
      </div>
    </button>
    <div style:grid-column="contextual-information" class="justify-self-end">
      <div class="italic text-gray-600">
        {#if selectingColumns}
          <span>
            {#if selectedColumns.length}
              selected {selectedColumns.length} column{#if selectedColumns.length > 1}s{/if}
            {:else}
              select columns
            {/if}
          </span>
        {:else}
          <span
            class="grid grid-flow-col gap-x-2 text-gray-500 text-clip overflow-hidden whitespace-nowrap "
          >
            {#if showEntityDetails}
              {#if showRows}
                <span>
                  <span>
                    {cardinality !== undefined && !isNaN(cardinality)
                      ? formatInteger(interimCardinality)
                      : "no"}
                  </span>
                  row{#if cardinality !== 1}s{/if}
                </span>
              {/if}
              <span class="self-center">
                <ContextButton
                  id={contextButtonId}
                  tooltipText=""
                  suppressTooltip={true}
                  on:click={clickContextButtonHandler}
                >
                  <MoreIcon />
                </ContextButton>
              </span>
              <slot />
            {/if}
          </span>
        {/if}
      </div>
    </div>
  </div>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        {name}
      </svelte:fragment>
      <svelte:fragment slot="description" />
    </TooltipTitle>
    <TooltipShortcutContainer>
      {#if entityType === EntityType.Table}
        <div>
          <StackingWord key="command">query</StackingWord> in workspace
        </div>
        <Shortcut>command + click</Shortcut>
      {/if}
      {#if entityType === EntityType.Model}
        <div>open in workspace</div>
        <Shortcut>click</Shortcut>
      {/if}
      <div>
        <StackingWord key="shift">copy</StackingWord> name to clipboard
      </div>
      <Shortcut>shift + click</Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
