<script lang="ts">
  import { Menu, MenuItem } from "$lib/components/menu";

  import { WithToggleableFloatingElement } from "$lib/components/floating-element";
  let item;
</script>

<WithToggleableFloatingElement
  let:active
  let:handleClose
  let:toggleFloatingElement
>
  <button on:click={toggleFloatingElement}>click me! {active} {item}</button>
  <Menu
    dark
    on:item-select={handleClose}
    on:lose-focus={handleClose}
    on:escape={handleClose}
    slot="floating-element"
  >
    <MenuItem
      on:select={() => {
        item = 1;
      }}>one</MenuItem
    >
    <MenuItem
      on:select={() => {
        item = 2;
      }}>two</MenuItem
    >
    <MenuItem
      on:select={() => {
        item = 3;
      }}>three</MenuItem
    >
  </Menu>
</WithToggleableFloatingElement>

<p>Here is more text underneath this! {item}</p>
