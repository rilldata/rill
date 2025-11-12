<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    createAdminServiceListProjectsForOrganizationAndUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let organization: string;
  export let userId: string;

  let isDropdownOpen = false;
  $: hasUserId = !!userId;

  $: userProjectsQuery = createAdminServiceListProjectsForOrganizationAndUser(
    organization,
    { userId },
    {
      query: {
        enabled: !!userId,
      },
    },
  );
  $: ({ data, isPending, error } = $userProjectsQuery);
  let projects: V1Project[];
  $: projects = data?.projects ?? [];

  function getProjectShareUrl(projectName: string) {
    // Link the user to the project dashboard list and open the share popover immediately.
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }
</script>

{#if hasUserId}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      <span class="capitalize">
        {projects.length} Project{projects.length > 1 ? "s" : ""}
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
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-gray-400">No projects</div>
{/if}
