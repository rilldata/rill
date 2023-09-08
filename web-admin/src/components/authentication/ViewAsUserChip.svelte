<script lang="ts">
  import { page } from "$app/stores";
  import {
    Popover,
    PopoverButton,
    PopoverPanel,
  } from "@rgossiaux/svelte-headlessui";
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { createPopperActions } from "svelte-popperjs";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "./viewAsUserStore";

  // Position the popover
  const [popperRef, popperContent] = createPopperActions();
  const popperOptions = {
    placement: "bottom-start",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };
</script>

<Popover use={[popperRef]} let:close let:open>
  <Chip
    removable
    on:remove={() => {
      // updateMimickedJWT(queryClient, null);
      viewAsUserStore.set(null);
    }}
    active={open}
  >
    <div slot="body">
      <PopoverButton>
        <div class="flex gap-x-2">
          <div>
            Viewing as <span class="font-bold">{$viewAsUserStore.email}</span>
          </div>
          <div class="flex items-center">
            <IconSpaceFixer pullRight>
              <div class="transition-transform" class:-rotate-180={open}>
                <CaretDownIcon size="14px" />
              </div>
            </IconSpaceFixer>
          </div>
        </div>
      </PopoverButton>
      <PopoverPanel use={[[popperContent, popperOptions]]}>
        <ViewAsUserPopover
          organization={$page.params.organization}
          project={$page.params.project}
          on:select={() => close(undefined)}
        />
      </PopoverPanel>
    </div>
    <svelte:fragment slot="remove-tooltip">
      <slot name="remove-tooltip-content">Clear view</slot>
    </svelte:fragment>
  </Chip>
</Popover>
