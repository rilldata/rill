<!-- web-admin/src/routes/-/admin/projects/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import {
    searchProjects,
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/admin/projects/selectors";

  let bannerRef: ActionResultBanner;
  let searchQuery = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

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
      try {
        await $hibernateProject.mutateAsync({ organization: org, project });
        bannerRef.show("success", `Project ${name} hibernated`);
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
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
      try {
        await $redeployProject.mutateAsync({ organization: org, project });
        bannerRef.show("success", `Project ${name} redeployed`);
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Projects"
  description="Search projects by name pattern, view details, hibernate or redeploy."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search projects (e.g. org/project, min 2 chars)..."
    on:search={handleSearch}
  />
</div>

{#if $projectsQuery.isLoading && searchQuery.length >= 2}
  <p class="text-sm text-slate-500">Searching...</p>
{:else if $projectsQuery.data?.names?.length}
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
              <button class="action-btn" on:click={() => handleHibernate(name)}>
                Hibernate
              </button>
              <button
                class="action-btn destructive"
                on:click={() => handleRedeploy(name)}
              >
                Redeploy
              </button>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else if searchQuery.length >= 2 && $projectsQuery.isSuccess}
  <p class="text-sm text-slate-500">No projects found for "{searchQuery}"</p>
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
  tr:hover td { @apply bg-slate-50 dark:bg-slate-800/50; }
  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700;
  }
  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }
</style>
