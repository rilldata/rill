<script lang="ts">
  import { page } from "$app/stores";
  import type { Modifier } from "@popperjs/core";
  import {
    Popover,
    PopoverButton,
    PopoverPanel,
  } from "@rgossiaux/svelte-headlessui";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { createEventDispatcher } from "svelte";
  import type { ContentAction } from "svelte-popperjs";
  import ViewAsUserPopover from "../../features/view-as-user/ViewAsUserPopover.svelte";

  export let popperContent: ContentAction<Partial<Modifier<any, any>>>;

  const popperOptions = {
    placement: "left-start",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };

  const dispatch = createEventDispatcher();
</script>

<Popover>
  <PopoverButton class="w-full text-left">
    <MenuItem animateSelect={false}>
      View as
      <CaretDownIcon
        className="transform -rotate-90"
        slot="right"
        size="14px"
      />
    </MenuItem>
  </PopoverButton>
  <PopoverPanel use={[[popperContent, popperOptions]]}>
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project}
      on:select={() => dispatch("select-user")}
    />
  </PopoverPanel>
</Popover>
