<!-- Form dialog for changing a project's prod slots. Loads the current value
     on open and commits via UpdateProject. -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { adminServiceGetProject } from "@rilldata/web-admin/client";
  import { createUpdateProjectMutation } from "@rilldata/web-admin/features/superuser/projects/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open = false;
  export let projectPath: string; // "org/project"

  const updateProject = createUpdateProjectMutation();

  let currentValue = "";
  let value = "";
  let loading = false;
  let saving = false;

  $: if (open && projectPath) {
    void loadCurrentSlots();
  }

  async function loadCurrentSlots() {
    loading = true;
    value = "";
    currentValue = "";
    try {
      const [org, project] = projectPath.split("/");
      const resp = await adminServiceGetProject(org, project, {
        superuserForceAccess: true,
      });
      currentValue = resp.project?.prodSlots ?? "";
      value = currentValue;
    } catch {
      currentValue = "?";
    } finally {
      loading = false;
    }
  }

  async function handleSave() {
    const slots = parseInt(value, 10);
    if (!slots || slots < 1) {
      eventBus.emit("notification", {
        type: "error",
        message: "Prod slots must be a positive integer",
      });
      return;
    }
    const [org, project] = projectPath.split("/");
    saving = true;
    try {
      await $updateProject.mutateAsync({
        org,
        project,
        data: { prodSlots: String(slots) },
      });
      eventBus.emit("notification", {
        type: "success",
        message: `Prod slots for ${projectPath} set to ${slots}`,
      });
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to update slots: ${err}`,
      });
    } finally {
      saving = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Change Prod Slots</AlertDialogTitle>
      <AlertDialogDescription>
        Update the number of production slots for {projectPath}.
      </AlertDialogDescription>
    </AlertDialogHeader>
    {#if loading}
      <p class="text-sm text-fg-secondary py-2">Loading current slots...</p>
    {:else}
      <div class="flex flex-col gap-3 py-2">
        <div class="flex items-center gap-2">
          <span class="text-sm text-fg-secondary">Current:</span>
          <span class="text-sm font-mono text-fg-primary"
            >{currentValue || "\u2014"}</span
          >
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-fg-secondary" for="slots-input"
            >New slots</label
          >
          <input
            id="slots-input"
            type="number"
            class="w-full px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            placeholder="Number of slots"
            min="1"
            bind:value
            on:keydown={(e) => e.key === "Enter" && handleSave()}
          />
        </div>
      </div>
    {/if}
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        onClick={handleSave}
        disabled={saving || loading || !value}
        loading={saving}
      >
        Save
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
