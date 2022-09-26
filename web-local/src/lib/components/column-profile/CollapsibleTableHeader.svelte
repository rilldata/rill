<script lang="ts">
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import ContextButton from "./ContextButton.svelte";
  import ExpanderButton from "./ExpanderButton.svelte";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import MoreIcon from "../icons/MoreHorizontal.svelte";
  import notificationStore from "./CollapsibleTableHeader.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import { createCommandClickAction } from "../../util/command-click-action";
  import { guidGenerator } from "../../util/guid";
  import { isMac } from "../../util/os-detection";
  import { createShiftClickAction } from "../../util/shift-click-action";
  import { format } from "d3-format";
  import { createEventDispatcher } from "svelte";
  import { cubicInOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu } from "../menu";

  export let entityType: EntityType;
  export let name: string;
  export let cardinality: number = undefined;
  export let showRows = true;
  export let sizeInBytes: number = undefined;
  export let active = false;
  export let show = false;
  export let contextMenuOpen = false;
  export let notExpandable = false;

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
      (entityType == EntityType.Table || entityType == EntityType.Model) &&
      active
    ) {
      show = !show;
    }
  };

  /** When the context or expander button is hovered, we should suppress the overall tooltip */
  let contextButtonIsHovered = false;
  let expanderIsHovered = false;
</script>

<Tooltip
  location="right"
  suppress={contextButtonIsHovered || expanderIsHovered || contextMenuOpen}
>
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
    class=" grid grid-flow-col gap-2 items-center hover:bg-gray-200 pl-4 pr-4 {active ||
    contextMenuOpen
      ? 'bg-gray-100'
      : 'bg-transparent'}
    "
  >
    {#if !notExpandable}
      <ExpanderButton
        bind:isHovered={expanderIsHovered}
        rotated={show}
        on:click={() => dispatch("expand")}
      >
        <CaretDownIcon size="14px" />
      </ExpanderButton>
    {:else}
      <Spacer size="16px" />
    {/if}
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
            {name.split(".").slice(0, -1).join(".")}
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
              <WithTogglableFloatingElement
                location="right"
                alignment="start"
                distance={16}
                let:toggleFloatingElement
                bind:active={contextMenuOpen}
              >
                <span class="self-center">
                  <ContextButton
                    id={contextButtonId}
                    tooltipText="more actions"
                    suppressTooltip={contextMenuOpen}
                    on:click={toggleFloatingElement}
                    bind:isHovered={contextButtonIsHovered}
                  >
                    <MoreIcon />
                  </ContextButton>
                </span>
                <Menu
                  dark
                  on:click-outside={toggleFloatingElement}
                  on:escape={toggleFloatingElement}
                  on:item-select={toggleFloatingElement}
                  slot="floating-element"
                >
                  <slot name="menu-items" toggleMenu={toggleFloatingElement} />
                </Menu>
              </WithTogglableFloatingElement>
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
        <Shortcut>
          {#if isMac()}<span
              style="
          font-family: var(--system); 
          font-size: 11.5px;
        ">⌘</span
            >{:else}ctrl{/if} + Click</Shortcut
        >
      {/if}
      {#if entityType === EntityType.Model}
        <div>open in workspace</div>
        <Shortcut>Click</Shortcut>
      {/if}
      <div>
        <StackingWord key="shift">copy</StackingWord> name to clipboard
      </div>
      <Shortcut>
        <span
          style="
          font-family: var(--system); 
          font-size: 11.5px;
        ">⇧</span
        > + Click</Shortcut
      >
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
