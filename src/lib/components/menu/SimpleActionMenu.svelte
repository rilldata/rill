<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { setContext } from "svelte";

  import Menu from "./Menu.svelte";
  import MenuItem from "./MenuItem.svelte";
  import WithFloatingMenu from "./WithFloatingMenu.svelte";

  export let options = [];
  export let dark: boolean = undefined;
  export let location: "left" | "right" | "top" | "bottom" = "bottom";
  export let alignment: "start" | "middle" | "end" = "start";
  export let distance = 16;

  export let active = false;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  function createOnClickHandler(callback, closeEventHandler) {
    return () => {
      callback();
      closeEventHandler();
    };
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
    on:select-item={handleClose}
    on:lose-focus={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { main, right, callback }, i}
      <MenuItem on:select={createOnClickHandler(callback, handleClose)}>
        {main}
        <svelte:fragment slot="right">
          {right || ""}
        </svelte:fragment>
      </MenuItem>
    {/each}
  </Menu>
</WithFloatingMenu>
