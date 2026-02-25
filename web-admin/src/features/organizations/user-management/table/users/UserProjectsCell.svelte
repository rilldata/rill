<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    createAdminServiceListProjectsForOrganizationAndUser,
    type V1UserProject,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { formatProjectRole } from "@rilldata/web-common/features/users/roles";

  export let organization: string;
  export let userId: string;
  export let projectCount: number;

  let isDropdownOpen = false;
  $: hasUserId = !!userId;

  $: userProjectsQuery = createAdminServiceListProjectsForOrganizationAndUser(
    organization,
    { userId },
    {
      query: {
        enabled: !!userId && isDropdownOpen,
      },
    },
  );
  $: ({ data, isPending, error } = $userProjectsQuery);
  let userProjects: V1UserProject[];
  $: userProjects = data?.projects ?? [];

  $: hasProjects = projectCount > 0;

  function getProjectShareUrl(projectName: string) {
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }
</script>

{#if hasUserId && hasProjects}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span class="capitalize">
        {projectCount} Project{projectCount !== 1 ? "s" : ""}
      </span>
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content align="start">
      {#if isPending}
        Loading...
      {:else if error}
        Error
      {:else}
        {#each userProjects as userProject (userProject.project?.id)}
          {@const projectName = userProject.project?.name ?? ""}
          {@const role = userProject.projectRoleName}
          <Dropdown.Item
            href={getProjectShareUrl(projectName)}
            class="flex items-center gap-2 whitespace-nowrap"
          >
            <span class="truncate">{projectName}</span>
            {#if role}
              <span class="text-fg-muted text-xs">
                {formatProjectRole(role)}
              </span>
            {/if}
          </Dropdown.Item>
        {/each}
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-fg-secondary">No projects</div>
{/if}
