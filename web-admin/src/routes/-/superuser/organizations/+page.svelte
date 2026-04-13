<script lang="ts">
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import ConfirmActionDialog from "@rilldata/web-admin/features/superuser/dialogs/ConfirmActionDialog.svelte";
  import GuardedDeleteDialog from "@rilldata/web-admin/features/superuser/dialogs/GuardedDeleteDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    getOrganization,
    getOrgMembers,
    getOrgProjects,
    createDeleteOrgMutation,
  } from "@rilldata/web-admin/features/superuser/organizations/selectors";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";

  let selectedOrg = "";

  // Open as User dialog state
  let assumeDialogOpen = false;
  let assumeEmail = "";
  let assumeOrgName = "";

  // Delete Org dialog state
  let deleteOrgDialogOpen = false;
  let deleteOrgName = "";
  let deleteOrgLoading = false;
  let deleteOrgError: string | undefined = undefined;

  const deleteOrg = createDeleteOrgMutation();

  // Load details for the selected org
  $: orgQuery = getOrganization(selectedOrg);
  $: membersQuery = getOrgMembers(selectedOrg);
  $: projectsQuery = getOrgProjects(selectedOrg);

  async function doAssume() {
    assumedUser.assume(assumeEmail, { redirect: `/${assumeOrgName}` });
  }

  async function doDeleteOrg() {
    deleteOrgLoading = true;
    deleteOrgError = undefined;
    try {
      await $deleteOrg.mutateAsync({ org: deleteOrgName });
      eventBus.emit("notification", {
        type: "success",
        message: `Organization "${deleteOrgName}" deleted`,
      });
      selectedOrg = "";
    } catch (err) {
      deleteOrgError = `Failed to delete organization: ${err}`;
      throw err;
    } finally {
      deleteOrgLoading = false;
    }
  }
</script>

<h1 class="text-lg font-semibold text-fg-primary">Organizations</h1>
<p class="text-sm text-fg-secondary mb-4">
  Search for any organization by name to view details, members, and projects.
</p>

<div class="mb-6 max-w-lg">
  <OrgPicker
    bind:value={selectedOrg}
    placeholder="Search organizations (min 3 characters)..."
  />
  {#if !selectedOrg}
    <p class="text-sm text-fg-muted mt-2">
      Type at least 3 characters to search by organization name.
    </p>
  {/if}
</div>

<!-- Selected org details -->
{#if selectedOrg}
  {#if $orgQuery.isFetching}
    <p class="text-sm text-fg-secondary py-4">Loading organization...</p>
  {:else if $orgQuery.isError}
    <p class="text-sm text-destructive">
      Organization "{selectedOrg}" not found or access denied.
    </p>
  {:else if $orgQuery.data?.organization}
    {@const org = $orgQuery.data.organization}
    <div class="flex flex-col gap-4">
      <section class="p-5 rounded-lg border">
        <h2 class="text-sm font-semibold text-fg-primary mb-3">
          Organization Details
        </h2>
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {#each [{ label: "Name", value: org.name }, { label: "Display Name", value: org.displayName ?? "-" }, { label: "Description", value: org.description ?? "-" }, { label: "Billing Plan", value: org.billingPlanDisplayName ?? "-" }, { label: "Billing Customer ID", value: org.billingCustomerId ?? "-", mono: true }, { label: "Custom Domain", value: org.customDomain ?? "None" }, { label: "Created", value: org.createdOn ? new Date(org.createdOn).toLocaleDateString() : "-" }, { label: "ID", value: org.id, mono: true }] as field}
            <div class="flex flex-col">
              <span class="text-sm text-fg-secondary uppercase tracking-wider"
                >{field.label}</span
              >
              <span class="text-sm text-fg-primary" class:font-mono={field.mono}
                >{field.value}</span
              >
            </div>
          {/each}
        </div>
        <div class="mt-4 pt-4 border-t">
          <Button
            large
            class="font-normal"
            type="secondary-destructive"
            disabled={!org.name}
            onClick={() => {
              deleteOrgName = org.name ?? "";
              deleteOrgDialogOpen = true;
            }}
          >
            Delete Organization
          </Button>
        </div>
      </section>

      <!-- Projects list (view-only; hibernate/redeploy are on the Projects page) -->
      {#if $projectsQuery.data?.projects?.length}
        <section class="p-5 rounded-lg border">
          <h2 class="text-sm font-semibold text-fg-primary mb-3">
            Projects ({$projectsQuery.data.projects.length})
          </h2>
          <div class="flex flex-col gap-1">
            {#each $projectsQuery.data.projects as project}
              <div
                class="flex items-center justify-between px-3 py-2 rounded bg-surface-subtle"
              >
                <Button
                  type="text"
                  onClick={() => {
                    const admin = $membersQuery.data?.members?.find(
                      (m) => m.roleName === "admin",
                    );
                    const member = admin ?? $membersQuery.data?.members?.[0];
                    if (member?.userEmail) {
                      assumedUser.assume(member.userEmail, {
                        redirect: `/${org.name}/${project.name}`,
                      });
                    }
                  }}
                >
                  {project.name}
                </Button>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      <!-- Members list -->
      {#if $membersQuery.isFetching}
        <p class="text-sm text-fg-secondary py-4">Loading members...</p>
      {:else if $membersQuery.data?.members?.length}
        <section class="p-5 rounded-lg border">
          <h2 class="text-sm font-semibold text-fg-primary mb-3">
            Members ({$membersQuery.data.members.length})
          </h2>
          <table class="w-full">
            <thead>
              <tr>
                <th
                  class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
                  >Email</th
                >
                <th
                  class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
                  >Role</th
                >
                <th
                  class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
                  >Actions</th
                >
              </tr>
            </thead>
            <tbody>
              {#each $membersQuery.data.members as member}
                <tr>
                  <td
                    class="px-4 py-2 text-sm font-mono text-fg-primary border-b"
                    >{member.userEmail}</td
                  >
                  <td class="px-4 py-2 text-sm text-fg-primary border-b"
                    >{member.roleName}</td
                  >
                  <td class="px-4 py-2 text-sm text-fg-primary border-b">
                    <Button
                      large
                      class="font-normal"
                      type="tertiary"
                      disabled={!member.userEmail}
                      onClick={() => {
                        assumeEmail = member.userEmail ?? "";
                        assumeOrgName = org.name ?? "";
                        assumeDialogOpen = true;
                      }}
                    >
                      Open as user
                    </Button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </section>
      {/if}
    </div>
  {/if}
{/if}

<ConfirmActionDialog
  bind:open={assumeDialogOpen}
  title="Open as User"
  description={`You will start browsing Rill Cloud as ${assumeEmail}, landing on the "${assumeOrgName}" organization. The session will expire after 60 minutes. Use the banner to unassume when done.`}
  onConfirm={doAssume}
/>

<GuardedDeleteDialog
  bind:open={deleteOrgDialogOpen}
  title="Delete Organization"
  description={`This will permanently delete "${deleteOrgName}" and all its projects, members, and data. This action cannot be undone.`}
  confirmText={deleteOrgName}
  confirmButtonText="Delete"
  loading={deleteOrgLoading}
  error={deleteOrgError}
  onConfirm={doDeleteOrg}
/>
