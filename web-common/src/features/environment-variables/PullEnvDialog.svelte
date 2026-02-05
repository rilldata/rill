<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import {
    createLocalServicePullEnv,
    createLocalServiceGetCurrentProject,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { get } from "svelte/store";

  export let open = false;
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

  let selectedEnvironment: "dev" | "prod" = "dev";

  const currentProjectQuery = createLocalServiceGetCurrentProject();
  const pullEnvMutation = createLocalServicePullEnv();

  $: isPending = $pullEnvMutation.isPending;
  $: error = $pullEnvMutation.error;

  async function handlePull() {
    const currentProject = get(currentProjectQuery).data;
    const project = currentProject?.project;

    // Try to get project info, but allow backend to infer if not available
    const orgName = project?.orgName ?? "";
    const projectName = project?.name ?? "";

    try {
      const result = await $pullEnvMutation.mutateAsync({
        org: orgName,
        project: projectName,
        environment: selectedEnvironment,
      });

      const variablesCount = result.variablesCount ?? 0;
      const modified = result.modified ?? false;

      if (!modified) {
        eventBus.emit("notification", {
          message:
            variablesCount === 0
              ? "No cloud credentials found for this project."
              : "Local .env file is already up to date with cloud credentials.",
        });
      } else {
        eventBus.emit("notification", {
          type: "success",
          message: `Successfully pulled ${variablesCount} variable${variablesCount === 1 ? "" : "s"} from Rill Cloud.`,
        });
      }

      open = false;
      onSuccess?.();
    } catch (err) {
      // Error is already handled by the mutation
      console.error("Failed to pull environment variables:", err);
    }
  }
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Pull Environment Variables</DialogTitle>
      <DialogDescription>
        Replace your local .env file with cloud variables.
      </DialogDescription>
    </DialogHeader>

    <div class="flex flex-col gap-y-2">
      <span class="text-sm font-medium text-fg-primary">Environment</span>
      <div class="flex gap-x-2">
        <button
          type="button"
          class="px-3 py-1.5 text-sm rounded-md border transition-colors {selectedEnvironment ===
          'dev'
            ? 'bg-primary-100 border-primary-500 text-primary-600'
            : 'bg-surface border text-fg-secondary hover:bg-surface-hover'}"
          on:click={() => (selectedEnvironment = "dev")}
        >
          Development
        </button>
        <button
          type="button"
          class="px-3 py-1.5 text-sm rounded-md border transition-colors {selectedEnvironment ===
          'prod'
            ? 'bg-primary-100 border-primary-500 text-primary-600'
            : 'bg-surface border text-fg-secondary hover:bg-surface-hover'}"
          on:click={() => (selectedEnvironment = "prod")}
        >
          Production
        </button>
      </div>
    </div>

    {#if !isProjectLinked}
      <p class="text-sm text-fg-muted">
        Deploy this project to Rill Cloud to sync environment variables.
      </p>
    {/if}

    {#if error}
      <div
        class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-800"
      >
        <p>{error?.message || "Failed to pull environment variables"}</p>
      </div>
    {/if}

    <DialogFooter>
      <Button type="tertiary" onClick={() => (open = false)} disabled={isPending}>
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={handlePull}
        disabled={isPending || !isProjectLinked}
        loading={isPending}
      >
        Pull from Rill Cloud
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
