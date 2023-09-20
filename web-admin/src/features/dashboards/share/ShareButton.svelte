<script lang="ts">
  import {
    Popover,
    PopoverButton,
    PopoverPanel,
  } from "@rgossiaux/svelte-headlessui";
  import { Button } from "@rilldata/web-common/components/button";
  import { Menu } from "@rilldata/web-common/components/menu";
  import MenuItem from "@rilldata/web-common/components/menu/core/MenuItem.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { createPopperActions } from "svelte-popperjs";

  // Position the Menu popover
  const [popperRef, popperContent] = createPopperActions();
  const popperOptions = {
    placement: "bottom-end",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };

  function handleCopyLink() {
    // Copy the current URL to the clipboard
    navigator.clipboard.writeText(window.location.href);

    notifications.send({
      message: "Link copied to clipboard",
    });
  }
</script>

<Popover class="relative">
  <PopoverButton use={[popperRef]}>
    <Button type="secondary">Share</Button>
  </PopoverButton>
  <PopoverPanel
    use={[[popperContent, popperOptions]]}
    class="max-w-fit absolute z-[1000]"
    let:close
  >
    <Menu minWidth="0px" focusOnMount={false} paddingBottom={0} paddingTop={0}>
      <MenuItem
        focusOnMount={false}
        on:select={() => {
          handleCopyLink();
          close(undefined);
        }}>Copy shareable link</MenuItem
      >
    </Menu>
  </PopoverPanel>
</Popover>
