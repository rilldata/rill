<script lang="ts">
  import AssumeUserDialog from "@rilldata/web-admin/features/superuser/dialogs/AssumeUserDialog.svelte";
  import DeleteOrgDialog from "@rilldata/web-admin/features/superuser/dialogs/DeleteOrgDialog.svelte";
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    getOrganization,
    getOrgMembers,
    getOrgProjects,
    pickAssumableMember,
  } from "@rilldata/web-admin/features/superuser/organizations/selectors";

  let selectedOrg = "";

  let assumeDialogOpen = false;
  let assumeEmail = "";
  let assumeRedirect: string | undefined = undefined;
  let assumeContextLabel = "";

  let deleteOrgDialogOpen = false;
  let deleteOrgName = "";

  $: orgQuery = getOrganization(selectedOrg);
  $: membersQuery = getOrgMembers(selectedOrg);
  $: projectsQuery = getOrgProjects(selectedOrg);

  function openAssume(
    email: string,
    redirect: string | undefined,
    contextLabel: string,
  ) {
    assumeEmail = email;
    assumeRedirect = redirect;
    assumeContextLabel = contextLabel;
    assumeDialogOpen = true;
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
      <section class="p-5 rounded-lg border">
        <h2 class="text-sm font-semibold text-fg-primary mb-3">
          Projects{$projectsQuery.data?.length
            ? ` (${$projectsQuery.data.length})`
            : ""}
        </h2>
        {#if $projectsQuery.isFetching}
          <p class="text-sm text-fg-secondary">Loading projects...</p>
        {:else if $projectsQuery.data?.length}
          <div class="flex flex-col gap-1">
            {#each $projectsQuery.data as projectName}
              <div
                class="flex items-center justify-between px-3 py-2 rounded bg-surface-subtle"
              >
                <Button
                  type="text"
                  onClick={() => {
                    const member = pickAssumableMember(
                      $membersQuery.data?.members,
                    );
                    if (!member) return;
                    openAssume(
                      member.userEmail,
                      `/${org.name}/${projectName}`,
                      `${org.name}/${projectName}`,
                    );
                  }}
                >
                  {projectName}
                </Button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="text-sm text-fg-secondary">
            No projects in this organization.
          </p>
        {/if}
      </section>

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
                      onClick={() =>
                        openAssume(
                          member.userEmail ?? "",
                          `/${org.name}`,
                          org.name ?? "",
                        )}
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

<AssumeUserDialog
  bind:open={assumeDialogOpen}
  email={assumeEmail}
  redirect={assumeRedirect}
  contextLabel={assumeContextLabel}
/>
<DeleteOrgDialog
  bind:open={deleteOrgDialogOpen}
  org={deleteOrgName}
  on:deleted={() => (selectedOrg = "")}
/>
