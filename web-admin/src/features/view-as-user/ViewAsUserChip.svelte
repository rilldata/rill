<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { errorStore } from "../../components/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import ViewAsUserOrgPopover from "./ViewAsUserOrgPopover.svelte";
  import {
    viewAsUserStore,
    viewAsUserStateStore$,
    clearViewAsUser,
  } from "./viewAsUserStore";

  export let isOrgAdmin: boolean = false;

  let active: boolean;

  // Determine if we should use org-level popover:
  // - When at org level (no project param) and user is org admin
  // - OR when the view-as was activated at org level (sourceProject is "__org_level__")
  $: isOrgLevelViewAs = $viewAsUserStateStore$?.sourceProject === "__org_level__";
  $: isAtOrgLevel = !$page.params.project;
  $: useOrgPopover = isOrgAdmin && (isAtOrgLevel || isOrgLevelViewAs);

  // Use the current project if available, otherwise fall back to the source project
  // where view-as was originally activated
  $: projectForUserQuery =
    $page.params.project ?? $viewAsUserStateStore$?.sourceProject;
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <Chip
      removable
      {active}
      builders={[builder]}
      removeTooltipText="Clear view"
      onRemove={() => {
        clearViewAsUser();
        errorStore.reset();
      }}
    >
      <div slot="body">
        Viewing as <b>{$viewAsUserStore?.email}</b>
      </div>
    </Chip>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class="flex flex-col min-w-[150px] max-w-[300px]"
  >
    {#if useOrgPopover}
      <ViewAsUserOrgPopover
        organization={$page.params.organization}
        onSelectUser={() => (active = false)}
      />
    {:else}
      <ViewAsUserPopover
        organization={$page.params.organization}
        project={projectForUserQuery}
        onSelectUser={() => (active = false)}
        isOrgLevel={isOrgAdmin}
      />
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
