<script lang="ts">
  import { page } from "$app/stores";
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { errorStore } from "../../features/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { clearViewedAsUserWithinProject } from "./clearViewedAsUser";
  import { viewAsUserStore } from "./viewAsUserStore";

  const queryClient = useQueryClient();

  $: org = $page.params.organization;
  $: project = $page.params.project;

  let active: boolean;
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger>
    <Chip
      removable
      on:remove={async () => {
        await clearViewedAsUserWithinProject(queryClient, org, project);
        errorStore.reset();
      }}
      {active}
    >
      <div slot="body">
        <div class="flex gap-x-2">
          <div>
            Viewing as <span class="font-bold">{$viewAsUserStore.email}</span>
          </div>
          <div class="flex items-center">
            <IconSpaceFixer pullRight>
              <div class="transition-transform" class:-rotate-180={active}>
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
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class="flex flex-col min-w-[150px] max-w-[300px] min-h-[150px] max-h-[190px]"
  >
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project}
      on:select={() => (active = false)}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
