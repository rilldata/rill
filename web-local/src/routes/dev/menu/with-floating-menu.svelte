<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";

  let item;
</script>

<WithTogglableFloatingElement
  let:active
  let:handleClose
  let:toggleFloatingElement
>
  <button on:click={toggleFloatingElement}>click me! {active} {item}</button>
  <Menu
    dark
    on:item-select={handleClose}
    on:click-outside={handleClose}
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
</WithTogglableFloatingElement>

<p>Here is more text underneath this! {item}</p>
