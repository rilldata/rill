<script lang="ts">
  import {
    Popover,
    PopoverButton,
    PopoverPanel,
  } from "@rgossiaux/svelte-headlessui";
  import { Button } from "@rilldata/web-common/components/button";
  import { Menu } from "@rilldata/web-common/components/menu";
  import MenuItem from "@rilldata/web-common/components/menu/core/MenuItem.svelte";
  import { createPopperActions } from "svelte-popperjs";

  // Position the Menu popover
  const [popperRef, popperContent] = createPopperActions();
  const popperOptions = {
    placement: "bottom-end",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };
</script>

<Popover class="relative">
  <PopoverButton use={[popperRef]}>
    <Button type="secondary">Share</Button>
  </PopoverButton>
  <PopoverPanel
    use={[[popperContent, popperOptions]]}
    class="max-w-fit absolute z-[1000]"
  >
    <Menu minWidth="0px" focusOnMount={false}>
      <MenuItem>Copy shareable link</MenuItem>
    </Menu>
  </PopoverPanel>
</Popover>
