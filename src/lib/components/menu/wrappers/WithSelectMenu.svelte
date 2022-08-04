<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import Check from "../../icons/Check.svelte";
  import Spacer from "../../icons/Spacer.svelte";

  import { WithTogglableFloatingElement } from "$lib/components/floating-element";
  import { Menu, MenuItem } from "../";

  export let options;
  export let selection = undefined;
  export let multiple = false;

  export let dark: boolean = undefined;
  export let location: "left" | "right" | "top" | "bottom" = "bottom";
  export let alignment: "start" | "middle" | "end" = "start";
  export let distance = 16;

  export let active = false;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  const dispatch = createEventDispatcher();

  let temporarilySelectedKey;
  function createOnClickHandler(
    main,
    right,
    description,
    key,
    index,
    closeEventHandler
  ) {
    return async () => {
      // single-select: do nothing if already selected
      if (isSelected(selection, key)) {
        return;
      }
      selection = { main, right, description, key, index };
      dispatch("select", selection);
      if (!multiple) closeEventHandler();

      temporarilySelectedKey = undefined;
    };
  }

  function isSelected(selection, key) {
    return selection.key === key;
  }
</script>

<WithTogglableFloatingElement
  bind:active
  {location}
  {alignment}
  {distance}
  let:handleClose
  let:toggleFloatingElement
>
  <slot {active} {handleClose} toggleMenu={toggleFloatingElement} />

  <Menu
    slot="floating-element"
    {dark}
    on:lose-focus={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { key, main, description, right }, i}
      {@const selected = isSelected(selection, key)}
      <MenuItem
        icon
        animateSelect={!multiple}
        on:before-select={() => {
          temporarilySelectedKey = key;
        }}
        on:select={createOnClickHandler(
          main,
          right,
          description,
          key,
          i,
          handleClose
        )}
        {selected}
      >
        <svelte:fragment slot="icon">
          <!-- this conditional will make the circle check appear briefly before the menu closes -->
          {#if (temporarilySelectedKey !== undefined && temporarilySelectedKey === key) || (temporarilySelectedKey === undefined && selected)}
            <Check />
          {:else}
            <Spacer />
          {/if}
        </svelte:fragment>

        {main}
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
