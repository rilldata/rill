<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import {
    createAdminServiceListProjectsForOrganizationAndUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let organization: string;
  export let userId: string;

  let isDropdownOpen = false;

  const userProjectsQuery =
    createAdminServiceListProjectsForOrganizationAndUser(organization, {
      userId,
    });
  $: ({ data, isPending, error } = $userProjectsQuery);
  let projects: V1Project[];
  $: projects = data?.projects ?? [];
</script>

<Popover.Root bind:open={isDropdownOpen}>
  <Popover.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <span class="capitalize"
      >{projects.length} Project{projects.length > 1 ? "s" : ""}</span
    >
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </Popover.Trigger>
  <Popover.Content>
    {#if isPending}
      Loading...
    {:else if error}
      Error
    {:else}
      <div class="flex flex-col gap-y-1">
        {#each projects as project (project.id)}
          <div class="text-gray-700 text-xs">{project.name}</div>
        {/each}
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
