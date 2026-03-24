<script lang="ts">
  import SuperuserPageHeader from "@rilldata/web-admin/features/superuser/layout/SuperuserPageHeader.svelte";
  import SearchInput from "@rilldata/web-admin/features/superuser/shared/SearchInput.svelte";
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
  import { adminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    searchProjects,
    createUpdateProjectMutation,
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/superuser/projects/selectors";

  let searchQuery = "";
  let dialogOpen = false;
  let dialogTitle = "";
  let dialogDescription = "";
  let dialogDestructive = false;
  let dialogAction: () => Promise<void> = async () => {};
  let dialogLoading = false;
  let actionInProgress = "";

  // Change Slots modal state
  let slotsDialogOpen = false;
  let slotsProjectName = "";
  let slotsValue = "";
  let slotsCurrentValue = "";
  let slotsLoading = false;
  let slotsSaving = false;

  const updateProject = createUpdateProjectMutation();
  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();

  $: projectsQuery = searchProjects(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  async function handleChangeSlots(name: string) {
    slotsProjectName = name;
    slotsValue = "";
    slotsCurrentValue = "";
    slotsLoading = true;
    slotsDialogOpen = true;
    try {
      const [org, project] = name.split("/");
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

  async function handleSaveSlots() {
    const slots = parseInt(slotsValue, 10);
    if (!slots || slots < 1) {
      eventBus.emit("notification", { type: "error", message: "Prod slots must be a positive integer" });
      return;
    }
    const [org, project] = slotsProjectName.split("/");
    slotsSaving = true;
    try {
      await $updateProject.mutateAsync({
        org,
        project,
        data: { prodSlots: String(slots), superuserForceAccess: true },
      });
      eventBus.emit("notification", { type: "success", message: `Prod slots for ${slotsProjectName} set to ${slots}` });
      slotsDialogOpen = false;
    } catch (err) {
      eventBus.emit("notification", { type: "error", message: `Failed to update slots: ${err}` });
    } finally {
      slotsSaving = false;
    }
  }

  function handleHibernate(name: string) {
    const [org, project] = name.split("/");
    dialogTitle = "Hibernate Project";
    dialogDescription = `This will hibernate the deployment for ${name}. The project data will be preserved but the deployment will be stopped.`;
    dialogDestructive = false;
    dialogAction = async () => {
      actionInProgress = `hibernate:${name}`;
      try {
        await $hibernateProject.mutateAsync({ org, project });
        eventBus.emit("notification", { type: "success", message: `Project ${name} hibernated` });
      } catch (err) {
        eventBus.emit("notification", { type: "error", message: `Failed: ${err}` });
      } finally {
        actionInProgress = "";
      }
    };
    dialogOpen = true;
  }

  function handleRedeploy(name: string) {
    const [org, project] = name.split("/");
    dialogTitle = "Redeploy Project";
    dialogDescription = `This will completely redeploy ${name}. This is a disruptive operation.`;
    dialogDestructive = true;
    dialogAction = async () => {
      actionInProgress = `redeploy:${name}`;
      try {
        await $redeployProject.mutateAsync({ org, project });
        eventBus.emit("notification", { type: "success", message: `Project ${name} redeployed` });
      } catch (err) {
        eventBus.emit("notification", { type: "error", message: `Failed: ${err}` });
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
  <p class="text-sm text-fg-secondary py-4">Searching projects...</p>
{:else if $projectsQuery.data?.names?.length}
  <p class="text-sm text-fg-secondary mb-2">
    {$projectsQuery.data.names.length} result{$projectsQuery.data.names.length === 1 ? "" : "s"}
  </p>
  <table class="w-full">
    <thead>
      <tr>
        <th class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b">Project</th>
        <th class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b">Actions</th>
      </tr>
    </thead>
    <tbody>
      {#each $projectsQuery.data.names as name}
        <tr>
          <td class="px-4 py-3 text-sm font-mono text-fg-primary border-b">{name}</td>
          <td class="px-4 py-3 text-sm text-fg-primary border-b">
            <div class="flex gap-2 items-center">
              <Button large class="font-normal" type="tertiary" href={`/${name}`} target="_blank">View</Button>
              <Button large class="font-normal" type="tertiary" onClick={() => handleChangeSlots(name)}>Change Slots</Button>
              <Button large class="font-normal"
                type="tertiary"
                               disabled={actionInProgress === `hibernate:${name}`}
                loading={actionInProgress === `hibernate:${name}`}
                onClick={() => handleHibernate(name)}
              >
                Hibernate
              </Button>
              <Button large class="font-normal"
                type="secondary-destructive"
                               disabled={actionInProgress === `redeploy:${name}`}
                loading={actionInProgress === `redeploy:${name}`}
                onClick={() => handleRedeploy(name)}
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
{:else if searchQuery.length < 3}
  <p class="text-sm text-fg-muted">
    Type at least 3 characters to search across all organizations.
  </p>
{/if}

<!-- Hibernate / Redeploy confirmation dialog -->
<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{dialogTitle}</AlertDialogTitle>
      <AlertDialogDescription>{dialogDescription}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button large class="font-normal" type="tertiary" onClick={() => (dialogOpen = false)}>Cancel</Button>
      <Button large class="font-normal"
        type={dialogDestructive ? "destructive" : "primary"}
        onClick={handleConfirm}
        loading={dialogLoading}
      >
        Confirm
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

<!-- Change Slots modal -->
<AlertDialog bind:open={slotsDialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Change Prod Slots</AlertDialogTitle>
      <AlertDialogDescription>
        Update the number of production slots for {slotsProjectName}.
      </AlertDialogDescription>
    </AlertDialogHeader>
    {#if slotsLoading}
      <p class="text-sm text-fg-secondary py-2">Loading current slots...</p>
    {:else}
      <div class="flex flex-col gap-3 py-2">
        <div class="flex items-center gap-2">
          <span class="text-sm text-fg-secondary">Current:</span>
          <span class="text-sm font-mono text-fg-primary">{slotsCurrentValue || "—"}</span>
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-fg-secondary" for="slots-input">New slots</label>
          <input
            id="slots-input"
            type="number"
            class="w-full px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            placeholder="Number of slots"
            min="1"
            bind:value={slotsValue}
            on:keydown={(e) => e.key === "Enter" && handleSaveSlots()}
          />
        </div>
      </div>
    {/if}
    <AlertDialogFooter>
      <Button large class="font-normal" type="tertiary" onClick={() => (slotsDialogOpen = false)}>Cancel</Button>
      <Button large class="font-normal"
        type="primary"
        onClick={handleSaveSlots}
        disabled={slotsSaving || slotsLoading || !slotsValue}
        loading={slotsSaving}
      >
        Save
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
