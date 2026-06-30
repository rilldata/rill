<script lang="ts">
  import AssumeUserDialog from "@rilldata/web-admin/features/superuser/dialogs/AssumeUserDialog.svelte";
  import ChangeSlotsDialog from "@rilldata/web-admin/features/superuser/dialogs/ChangeSlotsDialog.svelte";
  import HibernateProjectDialog from "@rilldata/web-admin/features/superuser/dialogs/HibernateProjectDialog.svelte";
  import RedeployProjectDialog from "@rilldata/web-admin/features/superuser/dialogs/RedeployProjectDialog.svelte";
  import SearchInput from "@rilldata/web-admin/features/superuser/shared/SearchInput.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { adminServiceListOrganizationMemberUsers } from "@rilldata/web-admin/client";
  import { pickAssumableMember } from "@rilldata/web-admin/features/superuser/organizations/selectors";
  import { searchProjects } from "@rilldata/web-admin/features/superuser/projects/selectors";

  let searchQuery = "";
  let viewInProgress = "";

  let assumeDialogOpen = false;
  let assumeEmail = "";
  let assumeRedirect: string | undefined = undefined;
  let assumeContextLabel = "";

  let hibernateDialogOpen = false;
  let redeployDialogOpen = false;
  let slotsDialogOpen = false;
  let targetProjectPath = "";

  $: projectsQuery = searchProjects(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  async function handleView(name: string) {
    const [org] = name.split("/");
    viewInProgress = name;
    try {
      const resp = await adminServiceListOrganizationMemberUsers(org, {
        superuserForceAccess: true,
      });
      const member = pickAssumableMember(resp.members);
      if (!member) {
        eventBus.emit("notification", {
          type: "error",
          message: `No members found in org "${org}" to assume as`,
        });
        return;
      }
      assumeEmail = member.userEmail;
      assumeRedirect = `/${name}`;
      assumeContextLabel = name;
      assumeDialogOpen = true;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to look up org members: ${err}`,
      });
    } finally {
      viewInProgress = "";
    }
  }
</script>

<h1 class="text-lg font-semibold text-fg-primary">Projects</h1>
<p class="text-sm text-fg-secondary mb-4">
  Search projects across all organizations. Change prod slots, hibernate, or
  redeploy.
</p>

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search projects (e.g. org/project, min 3 chars)..."
    on:search={handleSearch}
  />
  {#if searchQuery.length < 3}
    <p class="text-sm text-fg-muted mt-2">
      Type at least 3 characters to search across all organizations.
    </p>
  {/if}
</div>

{#if $projectsQuery.isFetching && searchQuery.length >= 3}
  <p class="text-sm text-fg-secondary py-4">Searching projects...</p>
{:else if $projectsQuery.data?.names?.length}
  <p class="text-sm text-fg-secondary mb-2">
    {$projectsQuery.data.names.length} result{$projectsQuery.data.names
      .length === 1
      ? ""
      : "s"}
  </p>
  <table class="w-full">
    <thead>
      <tr>
        <th
          class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
          >Project</th
        >
        <th
          class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
          >Actions</th
        >
      </tr>
    </thead>
    <tbody>
      {#each $projectsQuery.data.names as name}
        <tr>
          <td class="px-4 py-3 text-sm font-mono text-fg-primary border-b"
            >{name}</td
          >
          <td class="px-4 py-3 text-sm text-fg-primary border-b">
            <div class="flex gap-2 items-center">
              <Button
                large
                class="font-normal"
                type="tertiary"
                loading={viewInProgress === name}
                onClick={() => handleView(name)}>View</Button
              >
              <Button
                large
                class="font-normal"
                type="tertiary"
                onClick={() => {
                  targetProjectPath = name;
                  slotsDialogOpen = true;
                }}>Change Slots</Button
              >
              <Button
                large
                class="font-normal"
                type="tertiary"
                onClick={() => {
                  targetProjectPath = name;
                  hibernateDialogOpen = true;
                }}
              >
                Hibernate
              </Button>
              <Button
                large
                class="font-normal"
                type="secondary-destructive"
                onClick={() => {
                  targetProjectPath = name;
                  redeployDialogOpen = true;
                }}
              >
                Redeploy
              </Button>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else if searchQuery.length >= 3 && $projectsQuery.isSuccess}
  <p class="text-sm text-fg-secondary">No projects found for "{searchQuery}"</p>
{/if}

<AssumeUserDialog
  bind:open={assumeDialogOpen}
  email={assumeEmail}
  redirect={assumeRedirect}
  contextLabel={assumeContextLabel}
/>
<HibernateProjectDialog
  bind:open={hibernateDialogOpen}
  projectPath={targetProjectPath}
/>
<RedeployProjectDialog
  bind:open={redeployDialogOpen}
  projectPath={targetProjectPath}
/>
<ChangeSlotsDialog
  bind:open={slotsDialogOpen}
  projectPath={targetProjectPath}
/>
