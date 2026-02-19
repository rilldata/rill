<script lang="ts">
  import { onMount } from "svelte";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    adminServiceListProjectsForOrganization,
    adminServiceListProjectMemberUsergroups,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let organization: string;
  export let groupName: string;

  let isDropdownOpen = false;
  let isPending = true;
  let accessibleProjects: V1Project[] = [];
  let error: string | null = null;
  let hasLoaded = false;

  async function loadProjectsForGroup() {
    if (hasLoaded) return;

    isPending = true;
    error = null;

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
            const hasAccess = members.some((m) => m.groupName === groupName);
            return { project, hasAccess };
          } catch {
            return { project, hasAccess: false };
          }
        }),
      );

      accessibleProjects = projectAccessResults
        .filter((r) => r.hasAccess)
        .map((r) => r.project);
      hasLoaded = true;
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load projects";
    } finally {
      isPending = false;
    }
  }

  onMount(() => {
    void loadProjectsForGroup();
  });

  $: projectCount = accessibleProjects.length;
  $: hasProjects = projectCount > 0;

  function getProjectShareUrl(projectName: string | undefined) {
    return `/${organization}/${projectName}/-/dashboards?share=true`;
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
        <Dropdown.Item href={getProjectShareUrl(project.name)}>
          {project.name}
        </Dropdown.Item>
      {/each}
    </Dropdown.Content>
  </Dropdown.Root>
{:else if isPending}
  <div class="rounded-sm px-2 py-1 text-fg-secondary">Loading...</div>
{:else}
  <div class="rounded-sm px-2 py-1 text-fg-secondary">No projects</div>
{/if}
