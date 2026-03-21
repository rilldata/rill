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
  <div class="relative">
    <input
      type="text"
      class="w-full px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
      placeholder="Search organizations (min 3 characters)..."
      bind:value={searchInput}
      on:keydown={handleInputKeydown}
      on:input={handleInput}
      on:blur={handleBlur}
    />
    {#if $orgSearchQuery.isFetching && searchInput.length >= 3}
      <div
        class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
      />
    {/if}
    {#if showDropdown && orgNames.length > 0}
      <div
        class="absolute z-10 w-full mt-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-lg max-h-48 overflow-y-auto"
      >
        {#each orgNames as org}
          <button
            class="w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 cursor-pointer"
            on:mousedown|preventDefault={() => selectOrg(org)}
          >
            {org}
          </button>
        {/each}
      </div>
    {:else if showDropdown && searchInput.length >= 3 && $orgSearchQuery.isSuccess && orgNames.length === 0}
      <div
        class="absolute z-10 w-full mt-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-lg max-h-48 overflow-y-auto"
      >
        <div class="px-3 py-2 text-sm text-slate-400">No organizations found</div>
      </div>
    {/if}
  </div>
  <p class="text-xs text-slate-400 mt-1">
    Type to search, then select or press Enter for exact match.
  </p>
</div>

{#if $orgQuery.isFetching}
  <div class="flex items-center gap-2 py-4">
    <div
      class="w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
    />
    <span class="text-sm text-slate-500 dark:text-slate-400">Looking up organization...</span>
  </div>
{:else if $orgQuery.isError && lookupOrg}
  <p class="text-sm text-red-600">
    Organization "{lookupOrg}" not found or access denied.
  </p>
{:else if $orgQuery.data?.organization}
  {@const org = $orgQuery.data.organization}
  <div class="flex flex-col gap-4">
    <section class="p-5 rounded-lg border border-slate-200 dark:border-slate-700">
      <h2
        class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3"
      >
        Organization Details
      </h2>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >ID</span
          >
          <span
            class="text-sm text-slate-900 dark:text-slate-100 font-mono text-xs"
            >{org.id}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Name</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100"
            >{org.name}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Description</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100"
            >{org.description ?? "-"}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Billing Plan</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100"
            >{org.billingPlanDisplayName ?? "-"}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Billing Customer ID</span
          >
          <span
            class="text-sm text-slate-900 dark:text-slate-100 font-mono text-xs"
            >{org.billingCustomerId ?? "-"}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Custom Domain</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100"
            >{org.customDomain ?? "None"}</span
          >
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Created</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100">
            {org.createdOn
              ? new Date(org.createdOn).toLocaleDateString()
              : "-"}
          </span>
        </div>
        <div class="flex flex-col">
          <span
            class="text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider"
            >Projects</span
          >
          <span class="text-sm text-slate-900 dark:text-slate-100">
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
      <section
        class="p-5 rounded-lg border border-slate-200 dark:border-slate-700"
      >
        <h2
          class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3"
        >
          Projects ({$projectsQuery.data.projects.length})
        </h2>
        <div class="flex flex-col gap-1">
          {#each $projectsQuery.data.projects as project}
            <div
              class="flex items-center justify-between px-3 py-2 rounded bg-slate-50 dark:bg-slate-800"
            >
              <a
                href={`/${org.name}/${project.name}`}
                target="_blank"
                class="text-sm text-blue-600 dark:text-blue-400 hover:underline"
              >
                {project.name}
              </a>
              <div class="flex gap-2">
                <button
                  class="text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={actionInProgress ===
                    `hibernate:${project.name}`}
                  on:click={() =>
                    handleHibernate(org.name ?? "", project.name ?? "")}
                >
                  {actionInProgress === `hibernate:${project.name}`
                    ? "Hibernating..."
                    : "Hibernate"}
                </button>
                <button
                  class="text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20 disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={actionInProgress ===
                    `redeploy:${project.name}`}
                  on:click={() =>
                    handleRedeploy(org.name ?? "", project.name ?? "")}
                >
                  {actionInProgress === `redeploy:${project.name}`
                    ? "Redeploying..."
                    : "Redeploy"}
                </button>
              </div>
            </div>
          {/each}
        </div>
      </section>
    {/if}

    <!-- Members list -->
    {#if $membersQuery.isFetching}
      <div class="flex items-center gap-2 py-4">
        <div
          class="w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
        />
        <span class="text-sm text-slate-500 dark:text-slate-400">Loading members...</span>
      </div>
    {:else if $membersQuery.data?.members?.length}
      <section
        class="p-5 rounded-lg border border-slate-200 dark:border-slate-700"
      >
        <h2
          class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3"
        >
          Members ({$membersQuery.data.members.length})
        </h2>
        <table class="w-full">
          <thead>
            <tr>
              <th
                class="text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700"
              >
                Email
              </th>
              <th
                class="text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700"
              >
                Role
              </th>
            </tr>
          </thead>
          <tbody>
            {#each $membersQuery.data.members as member}
              <tr>
                <td
                  class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800 font-mono text-xs"
                >
                  {member.userEmail}
                </td>
                <td
                  class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800"
                >
                  {member.roleName}
                </td>
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
