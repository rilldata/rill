<script lang="ts">
  import { page } from "$app/stores";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { createAdminServiceGetService } from "@rilldata/web-admin/client";

  export let serviceName: string;
  export let hasProjectRoles: boolean;

  let isDropdownOpen = false;

  $: organization = $page.params.organization;
  $: serviceQuery = createAdminServiceGetService(organization, serviceName, {
    query: { enabled: hasProjectRoles },
  });
  $: memberships = $serviceQuery.data?.projectMemberships ?? [];
  $: projectCount = memberships.length;
</script>

{#if hasProjectRoles}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span>{projectCount} {projectCount === 1 ? "project" : "projects"}</span>
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content align="start">
      {#if $serviceQuery.isPending}
        <Dropdown.Item disabled>Loading...</Dropdown.Item>
      {:else if memberships.length > 0}
        {#each memberships as pm}
          <Dropdown.Item
            href="/{organization}/{pm.projectName}"
            class="flex justify-between gap-x-4"
          >
            <span>{pm.projectName}</span>
            <span class="text-fg-tertiary">{pm.projectRoleName}</span>
          </Dropdown.Item>
        {/each}
      {:else}
        <Dropdown.Item disabled>No projects</Dropdown.Item>
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <span class="text-fg-tertiary px-2 py-1">No projects</span>
{/if}
