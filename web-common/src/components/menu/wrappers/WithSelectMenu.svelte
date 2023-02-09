<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import { WithTogglableFloatingElement } from "../../floating-element";
  import type { Alignment, Location } from "../../floating-element/types";
  import Check from "../../icons/Check.svelte";
  import Spacer from "../../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../index";

  export let options;
  export let selection = undefined;

  export let dark: boolean = undefined;
  export let disabled: boolean = undefined;
  export let multiSelect: boolean = undefined;
  export let location: Location = "bottom";
  export let alignment: Alignment = "start";
  export let distance = 16;

  export let active = false;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  const dispatch = createEventDispatcher();

  let temporarilySelectedKey;
  function createOnClickHandler(
    main: string,
    right: string,
    description: string,
    key: string,
    disabled = false,
    index: number,
    closeEventHandler: () => void
  ) {
    return async () => {
      if (!multiSelect && isSelected(selection, key)) {
        return;
      }
      selection = { main, right, description, key, disabled, index };
      dispatch("select", selection);

      if (!multiSelect) closeEventHandler();

      temporarilySelectedKey = undefined;
    };
  }

  function isSelected(selection, key) {
    if (multiSelect) {
      return selection && selection.includes(key);
    }
    return selection === key || selection.key === key;
  }

  /** this function will make the circle check appear briefly before the menu closes */
  $: showCheckJustAfterClick = (key: string) =>
    temporarilySelectedKey !== undefined && temporarilySelectedKey === key;
  /** this function will otherwise render the check if selected, but only
   * if this is not part of the animation ticks
   */
  $: isAlreadySelectedButNotBeingAnimated = (
    key: string,
    isSelected: boolean
  ) => temporarilySelectedKey === undefined && isSelected;
</script>

<WithTogglableFloatingElement
  bind:active
  {location}
  {alignment}
  {distance}
  {disabled}
  let:handleClose
  let:toggleFloatingElement
>
  <slot {active} {handleClose} toggleMenu={toggleFloatingElement} />

  <Menu
    slot="floating-element"
    {dark}
    on:click-outside={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { key, main, description, right, disabled = false }, i}
      {@const selected = isSelected(selection, key)}
      <MenuItem
        icon
        animateSelect
        {disabled}
        on:before-select={() => {
          temporarilySelectedKey = key;
        }}
        on:select={createOnClickHandler(
          main,
          right,
          description,
          key,
          disabled,
          i,
          handleClose
        )}
        {selected}
      >
        <svelte:fragment slot="icon">
          {#if showCheckJustAfterClick(key) || isAlreadySelectedButNotBeingAnimated(key, selected)}
            <Check />
          {:else}
            <Spacer />
          {/if}
        </svelte:fragment>
        <div class:text-gray-400={disabled}>
          {main}
        </div>
        <svelte:fragment slot="description">
          {#if description}
            {description}
          {/if}
        </svelte:fragment>
        <svelte:fragment slot="right">
          {right || ""}
        </svelte:fragment>
      </MenuItem>
    {/each}
  </Menu>
</WithTogglableFloatingElement>
