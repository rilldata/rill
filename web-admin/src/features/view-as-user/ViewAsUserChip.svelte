<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { errorStore } from "../../components/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import {
    viewAsUserStore,
    viewAsUserStateStore$,
    clearViewAsUser,
  } from "./viewAsUserStore";

  export let isOrgAdmin: boolean = false;

  let active: boolean;
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
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project ?? $viewAsUserStateStore$?.projectContext}
      onSelectUser={() => (active = false)}
      isOrgLevel={isOrgAdmin}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
