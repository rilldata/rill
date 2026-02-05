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
  import type { EnvVariable } from "./types";
  import {
    createLocalServicePullEnv,
    createLocalServiceGetCurrentProject,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { get } from "svelte/store";

  export let open = false;
  export let currentVariables: EnvVariable[] = [];
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

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
        environment: "dev", // Default to dev environment
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
        Merge cloud variables with your local .env file. Existing variables will
        be updated with cloud values.
      </DialogDescription>
    </DialogHeader>

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
      <Button type="plain" onClick={() => (open = false)} disabled={isPending}>
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
