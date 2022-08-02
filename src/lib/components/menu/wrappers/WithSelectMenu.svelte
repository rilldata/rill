<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import Check from "../../icons/Check.svelte";
  import CheckBox from "../../icons/CheckBox.svelte";
  import EmptyBox from "../../icons/EmptyBox.svelte";
  import Spacer from "../../icons/Spacer.svelte";

  import { WithTogglableFloatingElement } from "$lib/components/floating-element";
  import { Menu, MenuItem } from "../";

  export let options;
  export let selections = [];
  /** set selections to the first option, if not provided on initialization */
  if (!selections.length) selections = [options[0]];
  export let style = "obvious";
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
      if (!multiple && isSelected(selections, key)) {
        return;
      }
      // set temporarily selected to get the icon to change instantly, then wait for tick
      // proceed with rest of update
      if (multiple) {
        // check to see if exists
        // if not, add.
        if (isSelected(selections, key)) {
          selections = [...selections.filter((s) => s.key !== key)];
        } else {
          selections = [
            ...selections,
            { main, right, key, description, index },
          ];
        }
      } else {
        // replace selected with a single value.
        selections = [{ main, right, description, key, index }];
      }
      dispatch("select", selections);
      if (!multiple) closeEventHandler();

      temporarilySelectedKey = undefined;
    };
  }

  function isSelected(selections, key) {
    return selections?.some((selection) => selection.key === key);
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
      {@const selected = isSelected(selections, key)}
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
          <!-- this conditional will make the circle check appear briefly before the menu closes
          in the case of a single-select menu. -->
          {#if !multiple}
            {#if (temporarilySelectedKey !== undefined && temporarilySelectedKey === key) || (temporarilySelectedKey === undefined && selected)}
              <!-- {#if style === "obvious"}<CheckCircle />{:else}<Check />{/if} -->
              <Check />
            {:else}
              <Spacer />
              <!-- {:else if style === "obvious"}
              <EmptyCircle />{:else}<Spacer /> -->
            {/if}
          {:else if selected}
            {#if style === "obvious"}<CheckBox />{:else}<Check />{/if}
          {:else if style === "obvious"}<EmptyBox />{:else}<Spacer />{/if}
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
