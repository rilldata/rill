<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import CheckBox from "../icons/CheckBox.svelte";
  import CheckCircle from "../icons/CheckCircle.svelte";
  import EmptyBox from "../icons/EmptyBox.svelte";
  import EmptyCircle from "../icons/EmptyCircle.svelte";

  import Menu from "./Menu.svelte";
  import MenuItem from "./MenuItem.svelte";
  import WithFloatingMenu from "./WithFloatingMenu.svelte";

  export let options;
  export let selections = [];
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
  function createOnClickHandler(main, right, key, index, closeEventHandler) {
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
          selections = [...selections, { main, right, key, index }];
        }
      } else {
        // replace selected with a single value.
        selections = [{ main, right, key, index }];
      }
      dispatch("select", { main, right, key, index });
      if (!multiple) closeEventHandler();

      temporarilySelectedKey = undefined;
    };
  }

  function isSelected(selections, key) {
    return selections?.some((selection) => selection.key === key);
  }
</script>

<WithFloatingMenu
  bind:active
  {location}
  {alignment}
  {distance}
  let:handleClose
  let:toggleMenu
>
  <slot {handleClose} {toggleMenu} {active} />
  <Menu
    slot="menu"
    {dark}
    on:lose-focus={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { key, main, right }, i}
      {@const selected = isSelected(selections, key)}
      <MenuItem
        icon
        animateSelect={!multiple}
        on:before-select={() => {
          temporarilySelectedKey = key;
        }}
        on:select={createOnClickHandler(main, right, key, i, handleClose)}
        {selected}
      >
        <svelte:fragment slot="icon">
          <!-- this conditional will make the circle check appear briefly before the menu closes
          in the case of a single-select menu. -->
          {#if !multiple}
            {#if (temporarilySelectedKey !== undefined && temporarilySelectedKey === key) || (temporarilySelectedKey === undefined && selected)}
              <CheckCircle />
            {:else}
              <EmptyCircle />
            {/if}
          {:else if selected}
            <CheckBox />
          {:else}
            <EmptyBox />
          {/if}
        </svelte:fragment>

        {main}
        <svelte:fragment slot="right">
          {right || ""}
        </svelte:fragment>
      </MenuItem>
    {/each}
  </Menu>
</WithFloatingMenu>
