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
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createPopperActions } from "svelte-popperjs";
  import { errorStore } from "../../components/errors/error-store";
  import { clearViewedAsUserWithinProject } from "./clearViewedAsUser";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "./viewAsUserStore";

  // Position the popover
  const [popperRef, popperContent] = createPopperActions();
  const popperOptions = {
    placement: "bottom-start",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };

  const queryClient = useQueryClient();
  $: org = $page.params.organization;
  $: project = $page.params.project;
</script>

<Popover let:close let:open>
  <PopoverButton use={[popperRef]}>
    <Chip
      removable
      on:remove={async () => {
        await clearViewedAsUserWithinProject(queryClient, org, project);
        errorStore.reset();
      }}
      active={open}
    >
      <div slot="body">
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
      </div>
      <svelte:fragment slot="remove-tooltip">
        <slot name="remove-tooltip-content">Clear view</slot>
      </svelte:fragment>
    </Chip>
  </PopoverButton>
  <PopoverPanel use={[[popperContent, popperOptions]]} class="z-[1000]">
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project}
      on:select={() => close(undefined)}
    />
  </PopoverPanel>
</Popover>
