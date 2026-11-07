<script lang="ts">
  import { page } from "$app/state";
  import { isEditPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { errorStore } from "../../components/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "./viewAsUserStore";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n/index.ts";

  let active: boolean = $state(false);
  let disabled = $derived(isEditPage(page));
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger {disabled}>
    {#snippet child({ props })}
      <Tooltip distance={8} suppress={!disabled}>
        <button
          {...props}
          {disabled}
          class="appearance-none border-0 bg-transparent p-0"
        >
          <Chip
            removable={!disabled}
            readOnly={disabled}
            {active}
            removeTooltipText={m.dashboard_clear_view()}
            onRemove={() => {
              viewAsUserStore.set(null);
              errorStore.reset();
            }}
          >
            <div slot="body">
              {@html m.dashboard_viewing_as({
                email: escapeHtml($viewAsUserStore?.email ?? ""),
              })}
            </div>
          </Chip>
        </button>
        <TooltipContent slot="tooltip-content">
          {m.dashboard_view_as_disabled()}
        </TooltipContent>
      </Tooltip>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class="flex flex-col min-w-[150px] max-w-[300px]"
  >
    <ViewAsUserPopover
      organization={page.params.organization}
      project={page.params.project}
      onSelectUser={() => (active = false)}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
