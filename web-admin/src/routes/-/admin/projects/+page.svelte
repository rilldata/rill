<!-- web-admin/src/routes/-/admin/projects/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import { adminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    searchProjects,
    createUpdateProjectMutation,
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

  // Change Slots inline edit state
  let slotsEditProject = "";
  let slotsValue = "";
  let slotsCurrentValue = "";
  let slotsLoading = false;

  const updateProject = createUpdateProjectMutation();
  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();

  $: projectsQuery = searchProjects(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  async function handleChangeSlots(name: string) {
    const [org, project] = name.split("/");
    slotsEditProject = name;
    slotsValue = "";
    slotsCurrentValue = "";
    slotsLoading = true;
    try {
      const resp = await adminServiceGetProject(org, project, {
        superuserForceAccess: true,
      });
      slotsCurrentValue = resp.project?.prodSlots ?? "";
      slotsValue = slotsCurrentValue;
    } catch {
      slotsCurrentValue = "?";
    } finally {
      slotsLoading = false;
    }
  }

  async function handleSaveSlots(name: string) {
    const slots = parseInt(slotsValue, 10);
    if (!slots || slots < 1) {
      notifyError("Prod slots must be a positive integer");
      return;
    }
    const [org, project] = name.split("/");
    actionInProgress = `slots:${name}`;
    try {
      await $updateProject.mutateAsync({
        org,
        project,
        data: { prodSlots: String(slots), superuserForceAccess: true },
      });
      notifySuccess(`Prod slots for ${name} set to ${slots}`);
      slotsEditProject = "";
      slotsValue = "";
    } catch (err) {
      notifyError(`Failed to update slots: ${err}`);
    } finally {
      actionInProgress = "";
    }
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
        notifySuccess(`Project ${name} hibernated`);
      } catch (err) {
        notifyError(`Failed: ${err}`);
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
        notifySuccess(`Project ${name} redeployed`);
      } catch (err) {
        notifyError(`Failed: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Projects"
  description="Search projects across all organizations. Change prod slots, hibernate, or redeploy."
/>

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search projects (e.g. org/project, min 3 chars)..."
    on:search={handleSearch}
  />
</div>

{#if $projectsQuery.isFetching && searchQuery.length >= 3}
  <div class="flex items-center gap-2 py-4">
    <div
      class="w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
    />
    <span class="text-sm text-slate-500">Searching projects...</span>
  </div>
{:else if $projectsQuery.data?.names?.length}
  <p class="text-xs text-slate-500 mb-2">
    {$projectsQuery.data.names.length} result{$projectsQuery.data.names
      .length === 1
      ? ""
      : "s"}
  </p>
  <table class="w-full">
    <thead>
      <tr>
        <th
          class="text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200"
        >
          Project
        </th>
        <th
          class="text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200"
        >
          Actions
        </th>
      </tr>
    </thead>
    <tbody>
      {#each $projectsQuery.data.names as name}
        <tr class="group">
          <td
            class="px-4 py-3 text-sm text-slate-700 border-b border-slate-100 group-hover:bg-slate-50 font-mono text-xs"
          >
            {name}
          </td>
          <td
            class="px-4 py-3 text-sm text-slate-700 border-b border-slate-100 group-hover:bg-slate-50"
          >
            <div class="flex gap-2 items-center">
              <a
                href={`/${name}`}
                target="_blank"
                class="text-xs px-2 py-1 rounded border border-slate-300 text-slate-600 hover:bg-slate-100"
              >
                View
              </a>
              {#if slotsEditProject === name}
                {#if slotsLoading}
                  <span class="text-xs text-slate-400">Loading...</span>
                {:else}
                  <span class="text-xs text-slate-500"
                    >Current: {slotsCurrentValue}</span
                  >
                  <input
                    type="number"
                    class="w-16 px-2 py-1 text-xs rounded border border-blue-400 bg-white text-slate-900 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    placeholder="slots"
                    min="1"
                    bind:value={slotsValue}
                    on:keydown={(e) => {
                      if (e.key === "Enter") handleSaveSlots(name);
                      if (e.key === "Escape") {
                        slotsEditProject = "";
                        slotsValue = "";
                      }
                    }}
                  />
                  <button
                    class="text-xs px-2 py-1 rounded border border-blue-400 text-blue-600 hover:bg-blue-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={actionInProgress === `slots:${name}` ||
                      !slotsValue}
                    on:click={() => handleSaveSlots(name)}
                  >
                    {actionInProgress === `slots:${name}`
                      ? "Saving..."
                      : "Save"}
                  </button>
                  <button
                    class="text-xs px-2 py-1 rounded border border-slate-300 text-slate-500 hover:bg-slate-100"
                    on:click={() => {
                      slotsEditProject = "";
                      slotsValue = "";
                    }}
                  >
                    Cancel
                  </button>
                {/if}
              {:else}
                <button
                  class="text-xs px-2 py-1 rounded border border-slate-300 text-slate-600 hover:bg-slate-100"
                  on:click={() => handleChangeSlots(name)}
                >
                  Change Slots
                </button>
              {/if}
              <button
                class="text-xs px-2 py-1 rounded border border-slate-300 text-slate-600 hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={actionInProgress === `hibernate:${name}`}
                on:click={() => handleHibernate(name)}
              >
                {actionInProgress === `hibernate:${name}`
                  ? "Hibernating..."
                  : "Hibernate"}
              </button>
              <button
                class="text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={actionInProgress === `redeploy:${name}`}
                on:click={() => handleRedeploy(name)}
              >
                {actionInProgress === `redeploy:${name}`
                  ? "Redeploying..."
                  : "Redeploy"}
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
