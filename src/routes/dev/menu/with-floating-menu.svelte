<script lang="ts">
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import WithFloatingMenu from "$lib/components/menu/WithFloatingMenu.svelte";

  let item;
</script>

<WithFloatingMenu let:active let:handleClose let:toggleMenu>
  <button on:click={toggleMenu}>click me! {active} {item}</button>
  <Menu
    dark
    on:item-select={handleClose}
    on:lose-focus={handleClose}
    on:escape={handleClose}
    slot="menu"
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
</WithFloatingMenu>

<p>Here is more text underneath this! {item}</p>
