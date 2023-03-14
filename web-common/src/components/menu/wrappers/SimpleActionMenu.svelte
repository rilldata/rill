<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { setContext } from "svelte";
  import { Menu, MenuItem } from "..";
  import { WithTogglableFloatingElement } from "../../floating-element";
  import type { Alignment, Location } from "../../floating-element/types";

  export let options = [];
  export let dark: boolean = undefined;
  export let location: Location = "bottom";
  export let alignment: Alignment = "start";
  export let distance = 16;
  export let minWidth = "300px";

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

<WithTogglableFloatingElement
  bind:active
  {location}
  {alignment}
  {distance}
  let:handleClose
  let:toggleFloatingElement
>
  <slot {handleClose} toggleMenu={toggleFloatingElement} {active} />
  <Menu
    slot="floating-element"
    {dark}
    {minWidth}
    focusOnMount={false}
    on:select-item={handleClose}
    on:click-outside={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { main, right, callback }, i}
      <MenuItem
        on:select={createOnClickHandler(callback, handleClose)}
        focusOnMount={false}
      >
        {main}
        <svelte:fragment slot="right">
          {right || ""}
        </svelte:fragment>
      </MenuItem>
    {/each}
  </Menu>
</WithTogglableFloatingElement>
