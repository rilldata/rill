<script lang="ts">
  import * as Command from "@rilldata/web-common/components/command/index.js";
  import {
    createAdminServiceListProjectsForOrganization,
    createAdminServiceSearchProjectUsers,
    type V1User,
  } from "../../client";
  import { errorStore } from "../../components/errors/error-store";
  import { setViewAsUser } from "./viewAsUserStore";

  export let organization: string;
  export let onSelectUser: (user: V1User) => void;

  let selectedProject: string | null = null;

  // Fetch projects for the organization
  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    { pageSize: 100 },
    {
      query: {
        enabled: !!organization,
      },
    },
  );

  // Fetch users for the selected project
  $: projectUsersQuery = createAdminServiceSearchProjectUsers(
    organization,
    selectedProject ?? "",
    { emailQuery: "%", pageSize: 1000, pageToken: undefined },
    {
      query: {
        enabled: !!organization && !!selectedProject,
      },
    },
  );

  $: projects = $projectsQuery.data?.projects ?? [];
  $: users = $projectUsersQuery.data?.users ?? [];

  function handleSelectProject(projectName: string) {
    selectedProject = projectName;
  }

  function handleViewAsUser(user: V1User) {
    if (!selectedProject) return;
    // Org-level view-as persists across projects
    setViewAsUser(user, selectedProject, true);
    errorStore.reset();
    onSelectUser(user);
  }

  function handleBack() {
    selectedProject = null;
  }
</script>

<div class="px-0.5 pt-1 pb-2 text-[10px] text-fg-secondary text-left">
  Test your <a
    target="_blank"
    href="https://docs.rilldata.com/build/metrics-view/security#rill-cloud"
    >security policies</a
  > by viewing this project from the perspective of another user.
</div>

{#if !selectedProject}
  <!-- Project Selection -->
  <Command.Root>
    <Command.Input placeholder="Search for a project" />
    <Command.List>
      <Command.Empty>No projects found.</Command.Empty>
      <Command.Group heading="Select a project">
        {#each projects as project}
          <Command.Item onSelect={() => handleSelectProject(project.name)}>
            {project.name}
          </Command.Item>
        {/each}
      </Command.Group>
    </Command.List>
  </Command.Root>
{:else}
  <!-- User Selection -->
  <button
    class="flex items-center gap-1 text-xs text-primary-500 hover:text-primary-600 mb-2 px-1"
    on:click={handleBack}
  >
    <svg
      class="w-3 h-3"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M15 19l-7-7 7-7"
      />
    </svg>
    Back to projects
  </button>
  <div class="text-xs text-fg-muted mb-2 px-1">
    Project: <span class="font-medium">{selectedProject}</span>
  </div>
  <Command.Root>
    <Command.Input placeholder="Search for users" />
    <Command.List>
      <Command.Empty>No users found.</Command.Empty>
      <Command.Group heading="Users">
        {#each users as user}
          <Command.Item onSelect={() => handleViewAsUser(user)}>
            {user.email}
          </Command.Item>
        {/each}
      </Command.Group>
    </Command.List>
  </Command.Root>
{/if}
