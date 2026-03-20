<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import {
    getOrganization,
    getOrgMembers,
    getOrgProjects,
    searchOrgNames,
  } from "@rilldata/web-admin/features/admin/organizations/selectors";
  import {
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/admin/projects/selectors";

  let searchInput = "";
  let lookupOrg = "";
  let showDropdown = false;
  let justSelected = false;
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};
  let actionInProgress = "";

  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();

  $: orgSearchQuery = searchOrgNames(searchInput);
  $: orgNames = extractUniqueOrgs($orgSearchQuery.data?.names ?? []);

  $: orgQuery = getOrganization(lookupOrg);
  $: membersQuery = getOrgMembers(lookupOrg);
  $: projectsQuery = getOrgProjects(lookupOrg);

  function extractUniqueOrgs(projectPaths: string[]): string[] {
    const orgs = new Set<string>();
    for (const path of projectPaths) {
      const org = path.split("/")[0];
      if (org) orgs.add(org);
    }
    return [...orgs].sort();
  }

  function selectOrg(org: string) {
    searchInput = org;
    lookupOrg = org;
    showDropdown = false;
    justSelected = true;
  }

  function handleInputKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && searchInput) {
      lookupOrg = searchInput;
      showDropdown = false;
      justSelected = true;
    }
  }

  function handleInput() {
    justSelected = false;
    if (searchInput.length >= 3) {
      showDropdown = true;
    } else {
      showDropdown = false;
    }
  }

  function handleBlur() {
    setTimeout(() => {
      showDropdown = false;
    }, 150);
  }

  function handleHibernate(orgName: string, projectName: string) {
    confirmTitle = "Hibernate Project";
    confirmDescription = `This will hibernate the deployment for ${orgName}/${projectName}. The project data will be preserved but the deployment will be stopped.`;
    confirmDestructive = false;
    confirmAction = async () => {
      actionInProgress = `hibernate:${projectName}`;
      try {
        await $hibernateProject.mutateAsync({ org: orgName, project: projectName });
        notifySuccess(`Project ${orgName}/${projectName} hibernated`);
      } catch (err) {
        notifyError(`Failed: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }

  function handleRedeploy(orgName: string, projectName: string) {
    confirmTitle = "Redeploy Project";
    confirmDescription = `This will completely redeploy ${orgName}/${projectName}. This is a disruptive operation.`;
    confirmDestructive = true;
    confirmAction = async () => {
      actionInProgress = `redeploy:${projectName}`;
      try {
        await $redeployProject.mutateAsync({ org: orgName, project: projectName });
        notifySuccess(`Project ${orgName}/${projectName} redeployed`);
      } catch (err) {
        notifyError(`Failed: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }

  $: if (orgNames.length > 0 && searchInput.length >= 3 && !justSelected) {
    showDropdown = true;
  }
</script>

<AdminPageHeader
  title="Organizations"
  description="Search for any organization by name to view details, members, and projects."
/>

<div class="mb-6 max-w-lg">
  <div class="search-container">
    <input
      type="text"
      class="input w-full"
      placeholder="Search organizations (min 3 characters)..."
      bind:value={searchInput}
      on:keydown={handleInputKeydown}
      on:input={handleInput}
      on:blur={handleBlur}
    />
    {#if $orgSearchQuery.isFetching && searchInput.length >= 3}
      <div class="search-spinner" />
    {/if}
    {#if showDropdown && orgNames.length > 0}
      <div class="dropdown">
        {#each orgNames as org}
          <button
            class="dropdown-item"
            on:mousedown|preventDefault={() => selectOrg(org)}
          >
            {org}
          </button>
        {/each}
      </div>
    {:else if showDropdown && searchInput.length >= 3 && $orgSearchQuery.isSuccess && orgNames.length === 0}
      <div class="dropdown">
        <div class="dropdown-empty">No organizations found</div>
      </div>
    {/if}
  </div>
  <p class="text-xs text-slate-400 mt-1">
    Type to search, then select or press Enter for exact match.
  </p>
</div>

{#if $orgQuery.isFetching}
  <div class="loading">
    <div class="spinner" />
    <span class="text-sm text-slate-500">Looking up organization...</span>
  </div>
{:else if $orgQuery.isError && lookupOrg}
  <p class="text-sm text-red-600">
    Organization "{lookupOrg}" not found or access denied.
  </p>
{:else if $orgQuery.data?.organization}
  {@const org = $orgQuery.data.organization}
  <div class="org-details">
    <section class="card">
      <h2 class="card-title">Organization Details</h2>
      <div class="detail-grid">
        <div class="detail-item">
          <span class="detail-label">ID</span>
          <span class="detail-value font-mono text-xs">{org.id}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Name</span>
          <span class="detail-value">{org.name}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Description</span>
          <span class="detail-value">{org.description ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Billing Plan</span>
          <span class="detail-value">{org.billingPlanDisplayName ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Billing Customer ID</span>
          <span class="detail-value font-mono text-xs"
            >{org.billingCustomerId ?? "-"}</span
          >
        </div>
        <div class="detail-item">
          <span class="detail-label">Custom Domain</span>
          <span class="detail-value">{org.customDomain ?? "None"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Created</span>
          <span class="detail-value">
            {org.createdOn
              ? new Date(org.createdOn).toLocaleDateString()
              : "-"}
          </span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Projects</span>
          <span class="detail-value">
            {#if $projectsQuery.isFetching}
              <span class="text-slate-400">Loading...</span>
            {:else if $projectsQuery.data?.projects}
              {$projectsQuery.data.projects.length}
            {:else}
              0
            {/if}
          </span>
        </div>
      </div>
    </section>

    <!-- Projects list -->
    {#if $projectsQuery.data?.projects?.length}
      <section class="card">
        <h2 class="card-title">
          Projects ({$projectsQuery.data.projects.length})
        </h2>
        <div class="project-list">
          {#each $projectsQuery.data.projects as project}
            <div class="project-row">
              <a
                href={`/${org.name}/${project.name}`}
                target="_blank"
                class="project-name"
              >
                {project.name}
              </a>
              <div class="flex gap-2">
                <button
                  class="action-btn"
                  disabled={actionInProgress === `hibernate:${project.name}`}
                  on:click={() => handleHibernate(org.name ?? "", project.name ?? "")}
                >
                  {actionInProgress === `hibernate:${project.name}` ? "Hibernating..." : "Hibernate"}
                </button>
                <button
                  class="action-btn destructive"
                  disabled={actionInProgress === `redeploy:${project.name}`}
                  on:click={() => handleRedeploy(org.name ?? "", project.name ?? "")}
                >
                  {actionInProgress === `redeploy:${project.name}` ? "Redeploying..." : "Redeploy"}
                </button>
              </div>
            </div>
          {/each}
        </div>
      </section>
    {/if}

    <!-- Members list -->
    {#if $membersQuery.isFetching}
      <div class="loading">
        <div class="spinner" />
        <span class="text-sm text-slate-500">Loading members...</span>
      </div>
    {:else if $membersQuery.data?.members?.length}
      <section class="card">
        <h2 class="card-title">
          Members ({$membersQuery.data.members.length})
        </h2>
        <table class="w-full">
          <thead>
            <tr>
              <th>Email</th>
              <th>Role</th>
            </tr>
          </thead>
          <tbody>
            {#each $membersQuery.data.members as member}
              <tr>
                <td class="font-mono text-xs">{member.userEmail}</td>
                <td>{member.roleName}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>
    {/if}
  </div>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  .search-container {
    @apply relative;
  }

  .search-spinner {
    @apply absolute right-3 top-1/2 -translate-y-1/2
      w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin;
  }

  .dropdown {
    @apply absolute z-10 w-full mt-1 bg-white dark:bg-slate-800
      border border-slate-200 dark:border-slate-700
      rounded-md shadow-lg max-h-48 overflow-y-auto;
  }

  .dropdown-item {
    @apply w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300
      hover:bg-slate-100 dark:hover:bg-slate-700 cursor-pointer;
  }

  .dropdown-empty {
    @apply px-3 py-2 text-sm text-slate-400;
  }

  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }

  .org-details {
    @apply flex flex-col gap-4;
  }

  .card {
    @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700;
  }

  .card-title {
    @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3;
  }

  .detail-grid {
    @apply grid grid-cols-2 lg:grid-cols-4 gap-3;
  }

  .detail-item {
    @apply flex flex-col;
  }

  .detail-label {
    @apply text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider;
  }

  .detail-value {
    @apply text-sm text-slate-900 dark:text-slate-100;
  }

  .project-list {
    @apply flex flex-col gap-1;
  }

  .project-row {
    @apply flex items-center justify-between px-3 py-2 rounded
      bg-slate-50 dark:bg-slate-800;
  }

  .project-name {
    @apply text-sm text-blue-600 dark:text-blue-400 hover:underline;
  }

  th {
    @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider
      px-4 py-2 border-b border-slate-200 dark:border-slate-700;
  }

  td {
    @apply px-4 py-2 text-sm text-slate-700 dark:text-slate-300
      border-b border-slate-100 dark:border-slate-800;
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
