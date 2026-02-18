<script lang="ts">
  import { onMount } from "svelte";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import {
    adminServiceListProjectsForOrganization,
    adminServiceListProjectMemberUsergroups,
    createAdminServiceAddProjectMemberUsergroup,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { Plus } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles";

  export let organization: string;
  export let groupName: string;

  const queryClient = useQueryClient();
  const addProjectMemberUsergroup = createAdminServiceAddProjectMemberUsergroup();

  let isDropdownOpen = false;
  let isAddProjectDropdownOpen = false;
  let isPending = true;
  let isAddingProject = false;
  let accessibleProjects: V1Project[] = [];
  let allProjects: V1Project[] = [];
  let error: string | null = null;
  let hasLoaded = false;

  async function loadProjectsForGroup() {
    if (hasLoaded) return;

    isPending = true;
    error = null;

    try {
      const projectsResponse =
        await adminServiceListProjectsForOrganization(organization);
      allProjects = projectsResponse.projects ?? [];

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

  async function handleAddProject(projectName: string) {
    isAddingProject = true;
    try {
      await $addProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: projectName,
        usergroup: groupName,
        data: {
          role: ProjectUserRoles.Viewer,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          projectName,
        ),
      });

      eventBus.emit("notification", {
        type: "success",
        message: `Added group "${groupName}" to project "${projectName}"`,
      });

      hasLoaded = false;
      await loadProjectsForGroup();
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to add group to project: ${e instanceof Error ? e.message : "Unknown error"}`,
      });
    } finally {
      isAddingProject = false;
      isAddProjectDropdownOpen = false;
    }
  }

  onMount(() => {
    void loadProjectsForGroup();
  });

  $: projectCount = accessibleProjects.length;
  $: hasProjects = projectCount > 0;
  $: availableProjectsToAdd = allProjects.filter(
    (p) => !accessibleProjects.some((ap) => ap.id === p.id),
  );

  function getProjectShareUrl(projectName: string | undefined) {
    return `/${organization}/${projectName}/-/dashboards?share=true`;
  }
</script>

<div class="flex items-center gap-1">
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

  {#if hasLoaded && availableProjectsToAdd.length > 0}
    <Dropdown.Root bind:open={isAddProjectDropdownOpen}>
      <Dropdown.Trigger
        class="flex items-center justify-center rounded-sm {isAddProjectDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} p-1"
        disabled={isAddingProject}
      >
        <Plus size="14px" class="text-fg-secondary" />
      </Dropdown.Trigger>
      <Dropdown.Content align="start" class="max-h-60 overflow-y-auto">
        <div class="px-2 py-1 text-xs text-fg-secondary font-medium">
          Add to project
        </div>
        {#each availableProjectsToAdd as project (project.id)}
          <Dropdown.Item
            on:click={() => handleAddProject(project.name ?? "")}
            disabled={isAddingProject}
          >
            {project.name}
          </Dropdown.Item>
        {/each}
      </Dropdown.Content>
    </Dropdown.Root>
  {/if}
</div>
