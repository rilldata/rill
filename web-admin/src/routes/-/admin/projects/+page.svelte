<!-- web-admin/src/routes/-/admin/projects/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import {
    searchProjects,
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/admin/projects/selectors";

  let searchQuery = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};
  let actionInProgress = "";

  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();

  $: projectsQuery = searchProjects(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleHibernate(name: string) {
    const [org, project] = name.split("/");
    confirmTitle = "Hibernate Project";
    confirmDescription = `This will hibernate the deployment for ${name}. The project data will be preserved but the deployment will be stopped.`;
    confirmDestructive = false;
    confirmAction = async () => {
      actionInProgress = `hibernate:${name}`;
      try {
        await $hibernateProject.mutateAsync({ org, project });
        notifySuccess( `Project ${name} hibernated`);
      } catch (err) {
        notifyError( `Failed: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }

  function handleRedeploy(name: string) {
    const [org, project] = name.split("/");
    confirmTitle = "Redeploy Project";
    confirmDescription = `This will completely redeploy ${name}. This is a disruptive operation.`;
    confirmDestructive = true;
    confirmAction = async () => {
      actionInProgress = `redeploy:${name}`;
      try {
        await $redeployProject.mutateAsync({ org, project });
        notifySuccess( `Project ${name} redeployed`);
      } catch (err) {
        notifyError( `Failed: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Projects"
  description="Search projects across all organizations by name pattern."
/>

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search projects (e.g. org/project, min 3 chars)..."
    on:search={handleSearch}
  />
</div>

{#if $projectsQuery.isFetching && searchQuery.length >= 3}
  <div class="loading">
    <div class="spinner" />
    <span class="text-sm text-slate-500">Searching projects...</span>
  </div>
{:else if $projectsQuery.data?.names?.length}
  <p class="text-xs text-slate-500 mb-2">
    {$projectsQuery.data.names.length} result{$projectsQuery.data.names.length === 1 ? "" : "s"}
  </p>
  <table class="w-full">
    <thead>
      <tr>
        <th>Project</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody>
      {#each $projectsQuery.data.names as name}
        <tr>
          <td class="font-mono text-xs">{name}</td>
          <td>
            <div class="flex gap-2">
              <a
                href={`/${name}`}
                target="_blank"
                class="action-btn"
              >
                View
              </a>
              <button
                class="action-btn"
                disabled={actionInProgress === `hibernate:${name}`}
                on:click={() => handleHibernate(name)}
              >
                {actionInProgress === `hibernate:${name}` ? "Hibernating..." : "Hibernate"}
              </button>
              <button
                class="action-btn destructive"
                disabled={actionInProgress === `redeploy:${name}`}
                on:click={() => handleRedeploy(name)}
              >
                {actionInProgress === `redeploy:${name}` ? "Redeploying..." : "Redeploy"}
              </button>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else if searchQuery.length >= 3 && $projectsQuery.isSuccess}
  <p class="text-sm text-slate-500">No projects found for "{searchQuery}"</p>
{:else if searchQuery.length < 3}
  <p class="text-sm text-slate-400">
    Type at least 3 characters to search across all organizations.
  </p>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  th {
    @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider
      px-4 py-2 border-b border-slate-200 dark:border-slate-700;
  }

  td {
    @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300
      border-b border-slate-100 dark:border-slate-800;
  }

  tr:hover td {
    @apply bg-slate-50 dark:bg-slate-800/50;
  }

  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700;
  }

  .action-btn:disabled {
    @apply opacity-50 cursor-not-allowed;
  }

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }

  .loading {
    @apply flex items-center gap-2 py-4;
  }

  .spinner {
    @apply w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin;
  }
</style>
