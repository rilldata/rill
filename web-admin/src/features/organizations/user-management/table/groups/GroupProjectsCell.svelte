<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListProjectMemberUsergroupsQueryOptions,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let organization: string;
  export let groupName: string;

  interface ProjectWithRole {
    id: string;
    name: string;
    roleName: string;
  }

  let isDropdownOpen = false;
  let accessibleProjects: ProjectWithRole[] = [];
  let isProcessing = false;
  let hasProcessed = false;

  const queryClient = useQueryClient();

  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: isDropdownOpen && !!organization,
      },
    },
  );

  $: allProjects = $projectsQuery?.data?.projects ?? ([] as V1Project[]);

  async function processProjectAccess(projects: V1Project[]) {
    if (hasProcessed || isProcessing || projects.length === 0) return;

    isProcessing = true;

    const results: ProjectWithRole[] = [];

    for (const project of projects) {
      try {
        const queryOptions =
          getAdminServiceListProjectMemberUsergroupsQueryOptions(
            organization,
            project.name ?? "",
          );

        const response = await queryClient.fetchQuery(queryOptions);
        const members = response?.members ?? [];
        const groupMember = members.find((m) => m.groupName === groupName);

        if (groupMember) {
          results.push({
            id: project.id ?? "",
            name: project.name ?? "",
            roleName: groupMember.roleName ?? "",
          });
        }
      } catch {
        // Ignore errors for individual projects
      }
    }

    accessibleProjects = results;
    hasProcessed = true;
    isProcessing = false;
  }

  $: if (isDropdownOpen && allProjects.length > 0 && !hasProcessed) {
    void processProjectAccess(allProjects);
  }

  $: projectCount = accessibleProjects.length;
  $: hasProjects = projectCount > 0;
  $: isPending = $projectsQuery?.isPending || isProcessing;

  function getProjectUrl(projectName: string) {
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }

  function formatRoleName(roleName: string): string {
    return roleName.charAt(0).toUpperCase() + roleName.slice(1).toLowerCase();
  }
</script>

<Dropdown.Root bind:open={isDropdownOpen}>
  <Dropdown.Trigger
    class="flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
      ? 'bg-gray-200'
      : 'hover:bg-surface-hover'} px-2 py-1"
  >
    <span class="capitalize">
      {#if hasProcessed}
        {projectCount} Project{projectCount !== 1 ? "s" : ""}
      {:else}
        Projects
      {/if}
    </span>
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </Dropdown.Trigger>
  <Dropdown.Content align="start">
    {#if isPending}
      <div class="px-3 py-2 text-fg-secondary">Loading...</div>
    {:else if !hasProjects}
      <div class="px-3 py-2 text-fg-secondary">No projects</div>
    {:else}
      {#each accessibleProjects as project (project.id)}
        <Dropdown.Item
          href={getProjectUrl(project.name)}
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center justify-between gap-4"
        >
          <span class="truncate">{project.name}</span>
          <span class="text-fg-secondary text-xs shrink-0"
            >{formatRoleName(project.roleName)}</span
          >
        </Dropdown.Item>
      {/each}
    {/if}
  </Dropdown.Content>
</Dropdown.Root>
