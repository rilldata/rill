<script lang="ts">
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { UserRoundSearch } from "lucide-svelte";
  import { errorStore } from "../../components/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "./viewAsUserStore";

  let open = false;
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      {#if $viewAsUserStore}
        <button
          {...props}
          class="appearance-none border-0 bg-transparent p-0"
          aria-label={`Viewing as ${$viewAsUserStore.email}`}
        >
          <Chip
            removable
            active={open}
            removeTooltipText="Clear view"
            onRemove={() => {
              viewAsUserStore.set(null);
              errorStore.reset();
            }}
          >
            <div slot="body">
              Viewing as <b>{$viewAsUserStore.email}</b>
            </div>
          </Chip>
        </button>
      {:else}
        <button
          {...props}
          type="button"
          class="flex items-center gap-x-2 px-3 h-7 text-primary-600 text-xs font-medium hover:bg-primary-50 transition-colors"
          aria-label="View as another user"
        >
          <UserRoundSearch size={14} />
          <span>View as</span>
        </button>
      {/if}
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="end"
    class="flex flex-col min-w-[220px] max-w-[320px]"
  >
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project}
      onSelectUser={() => (open = false)}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
