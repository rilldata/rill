<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    createAdminServiceListProjectsForOrganization,
    createAdminServiceListProjectsForOrganizationAndUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";

  export let organization: string;
  export let userId: string;
  export let role: string;

  let isDropdownOpen = false;

  const allProjectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  const userProjectsQuery =
    createAdminServiceListProjectsForOrganizationAndUser(organization, {
      userId,
    });
  $: ({ data, isPending, error } = $userProjectsQuery);
  let projects: V1Project[];
  $: projects = data?.projects ?? [];
  $: accessToAllProjects =
    $allProjectsQuery.data?.projects.length === projects.length;

  $: isGuest = role === OrgUserRoles.Guest;
  $: showAllProjects = !isGuest && accessToAllProjects;

  function getProjectShareUrl(projectName: string) {
    // Link the user to the project dashboard list and open the share popover immediately.
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }
</script>

<Dropdown.Root bind:open={isDropdownOpen}>
  <Dropdown.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <span class="capitalize">
      {#if showAllProjects}
        All projects
      {:else}
        {projects.length} Project{projects.length > 1 ? "s" : ""}
      {/if}
    </span>
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </Dropdown.Trigger>
  <Dropdown.Content>
    {#if isPending}
      Loading...
    {:else if error}
      Error
    {:else}
      {#each projects as project (project.id)}
        <Dropdown.Item href={getProjectShareUrl(project.name)}>
          {project.name}
        </Dropdown.Item>
      {/each}
    {/if}
  </Dropdown.Content>
</Dropdown.Root>
