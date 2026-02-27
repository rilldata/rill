<script lang="ts">
  import { onMount } from "svelte";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    adminServiceListProjectsForOrganization,
    adminServiceListProjectMemberUsergroups,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let organization: string;
  export let groupName: string;

  interface ProjectWithRole {
    id: string;
    name: string;
    roleName: string;
  }

  let isDropdownOpen = false;
  let isPending = true;
  let accessibleProjects: ProjectWithRole[] = [];
  let hasLoaded = false;

  async function loadProjectsForGroup() {
    if (hasLoaded) return;

    isPending = true;

    try {
      const projectsResponse =
        await adminServiceListProjectsForOrganization(organization);
      const allProjects = projectsResponse.projects ?? [];

      const projectAccessResults = await Promise.all(
        allProjects.map(async (project) => {
          try {
            const usergroupsResponse =
              await adminServiceListProjectMemberUsergroups(
                organization,
                project.name ?? "",
              );
            const members = usergroupsResponse.members ?? [];
            const groupMember = members.find((m) => m.groupName === groupName);
            if (groupMember) {
              return {
                project: {
                  id: project.id ?? "",
                  name: project.name ?? "",
                  roleName: groupMember.roleName ?? "",
                },
                hasAccess: true,
              };
            }
            return { project: null, hasAccess: false };
          } catch {
            return { project: null, hasAccess: false };
          }
        }),
      );

      accessibleProjects = projectAccessResults
        .filter((r) => r.hasAccess && r.project)
        .map((r) => r.project as ProjectWithRole);
      hasLoaded = true;
    } catch {
      // Ignore errors, will show "No projects"
    } finally {
      isPending = false;
    }
  }

  onMount(() => {
    void loadProjectsForGroup();
  });

  $: projectCount = accessibleProjects.length;
  $: hasProjects = projectCount > 0;

  function getProjectUrl(projectName: string) {
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }

  function formatRoleName(roleName: string): string {
    return roleName.charAt(0).toUpperCase() + roleName.slice(1).toLowerCase();
  }
</script>

{#if hasLoaded && hasProjects}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
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
    </Dropdown.Content>
  </Dropdown.Root>
{:else if isPending}
  <div class="rounded-sm px-2 py-1 text-fg-secondary">Loading...</div>
{:else}
  <div class="rounded-sm px-2 py-1 text-fg-secondary">No projects</div>
{/if}
