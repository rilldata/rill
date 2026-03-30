<script lang="ts">
  import SuperuserPageHeader from "@rilldata/web-admin/features/superuser/layout/SuperuserPageHeader.svelte";
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    getOrganization,
    getOrgMembers,
    getOrgProjects,
    createDeleteOrgMutation,
  } from "@rilldata/web-admin/features/superuser/organizations/selectors";
  import {
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/superuser/projects/selectors";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";

  let selectedOrg = "";
  let dialogOpen = false;
  let dialogTitle = "";
  let dialogDescription = "";
  let dialogDestructive = false;
  let dialogAction: () => Promise<void> = async () => {};
  let dialogLoading = false;
  let actionInProgress = "";

  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();
  const deleteOrg = createDeleteOrgMutation();

  // Load details for the selected org
  $: orgQuery = getOrganization(selectedOrg);
  $: membersQuery = getOrgMembers(selectedOrg);
  $: projectsQuery = getOrgProjects(selectedOrg);

  function handleOpenAsUser(email: string, orgName: string) {
    dialogTitle = "Open as User";
    dialogDescription = `You will start browsing Rill Cloud as ${email}, landing on the "${orgName}" organization. The session will expire after 60 minutes. Use the banner to unassume when done.`;
    dialogDestructive = false;
    dialogAction = async () => {
      assumedUser.assume(email, { redirect: `/${orgName}` });
    };
    dialogOpen = true;
  }

  function handleDeleteOrg(orgName: string) {
    dialogTitle = "Delete Organization";
    dialogDescription = `This will permanently delete "${orgName}" and all its projects, members, and data. This action cannot be undone.`;
    dialogDestructive = true;
    dialogAction = async () => {
      try {
        await $deleteOrg.mutateAsync({ org: orgName });
        eventBus.emit("notification", {
          type: "success",
          message: `Organization "${orgName}" deleted`,
        });
        selectedOrg = "";
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to delete organization: ${err}`,
        });
      }
    };
    dialogOpen = true;
  }

  function handleHibernate(orgName: string, projectName: string) {
    dialogTitle = "Hibernate Project";
    dialogDescription = `This will hibernate the deployment for ${orgName}/${projectName}. The project data will be preserved but the deployment will be stopped.`;
    dialogDestructive = false;
    dialogAction = async () => {
      actionInProgress = `hibernate:${projectName}`;
      try {
        await $hibernateProject.mutateAsync({
          org: orgName,
          project: projectName,
        });
        eventBus.emit("notification", {
          type: "success",
          message: `Project ${orgName}/${projectName} hibernated`,
        });
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed: ${err}`,
        });
      } finally {
        actionInProgress = "";
      }
    };
    dialogOpen = true;
  }

  function handleRedeploy(orgName: string, projectName: string) {
    dialogTitle = "Redeploy Project";
    dialogDescription = `This will completely redeploy ${orgName}/${projectName}. This is a disruptive operation.`;
    dialogDestructive = true;
    dialogAction = async () => {
      actionInProgress = `redeploy:${projectName}`;
      try {
        await $redeployProject.mutateAsync({
          org: orgName,
          project: projectName,
        });
        eventBus.emit("notification", {
          type: "success",
          message: `Project ${orgName}/${projectName} redeployed`,
        });
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed: ${err}`,
        });
      } finally {
        actionInProgress = "";
      }
    };
    dialogOpen = true;
  }

  async function handleConfirm() {
    dialogLoading = true;
    try {
      await dialogAction();
      dialogOpen = false;
    } catch {
      // Keep open for retry
    } finally {
      dialogLoading = false;
    }
  }
</script>

<SuperuserPageHeader
  title="Organizations"
  description="Search for any organization by name to view details, members, and projects."
/>

<div class="mb-6 max-w-lg">
  <OrgPicker
    bind:value={selectedOrg}
    placeholder="Search organizations (min 3 characters)..."
  />
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
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-fg-primary">
            Organization Details
          </h2>
          <Button
            large
            class="font-normal"
            type="destructive"
            onClick={() => handleDeleteOrg(org.name ?? "")}
          >
            Delete Organization
          </Button>
        </div>
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {#each [{ label: "ID", value: org.id, mono: true }, { label: "Name", value: org.name }, { label: "Display Name", value: org.displayName ?? "-" }, { label: "Description", value: org.description ?? "-" }, { label: "Billing Plan", value: org.billingPlanDisplayName ?? "-" }, { label: "Billing Customer ID", value: org.billingCustomerId ?? "-", mono: true }, { label: "Custom Domain", value: org.customDomain ?? "None" }, { label: "Created", value: org.createdOn ? new Date(org.createdOn).toLocaleDateString() : "-" }] as field}
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
      </section>

      <!-- Projects list -->
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
                <a
                  href={`/${org.name}/${project.name}`}
                  target="_blank"
                  class="text-sm text-accent-primary-action hover:underline"
                >
                  {project.name}
                </a>
                <div class="flex gap-2">
                  <Button
                    large
                    class="font-normal"
                    type="tertiary"
                    disabled={actionInProgress === `hibernate:${project.name}`}
                    loading={actionInProgress === `hibernate:${project.name}`}
                    onClick={() =>
                      handleHibernate(org.name ?? "", project.name ?? "")}
                  >
                    Hibernate
                  </Button>
                  <Button
                    large
                    class="font-normal"
                    type="secondary-destructive"
                    disabled={actionInProgress === `redeploy:${project.name}`}
                    loading={actionInProgress === `redeploy:${project.name}`}
                    onClick={() =>
                      handleRedeploy(org.name ?? "", project.name ?? "")}
                  >
                    Redeploy
                  </Button>
                </div>
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
                      onClick={() =>
                        handleOpenAsUser(
                          member.userEmail ?? "",
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

<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{dialogTitle}</AlertDialogTitle>
      <AlertDialogDescription>{dialogDescription}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        large
        class="font-normal"
        type="tertiary"
        onClick={() => (dialogOpen = false)}>Cancel</Button
      >
      <Button
        large
        class="font-normal"
        type={dialogDestructive ? "destructive" : "primary"}
        onClick={handleConfirm}
        loading={dialogLoading}
      >
        Confirm
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
