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
    createLocalServicePushEnv,
    createLocalServiceGetCurrentProject,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { get } from "svelte/store";

  export let open = false;
  export let currentVariables: EnvVariable[] = [];
  export let onSuccess: (() => void) | undefined = undefined;

  const currentProjectQuery = createLocalServiceGetCurrentProject();
  const pushEnvMutation = createLocalServicePushEnv();

  $: isPending = $pushEnvMutation.isPending;
  $: error = $pushEnvMutation.error;

  async function handlePush() {
    const currentProject = get(currentProjectQuery).data;
    const project = currentProject?.project;

    // Try to get project info, but allow backend to infer if not available
    const orgName = project?.orgName ?? "";
    const projectName = project?.name ?? "";

    try {
      const result = await $pushEnvMutation.mutateAsync({
        org: orgName,
        project: projectName,
        environment: "", // Empty for both environments
      });

      const addedCount = result.addedCount ?? 0;
      const changedCount = result.changedCount ?? 0;

      if (addedCount === 0 && changedCount === 0) {
        eventBus.emit("notification", {
          message: "No changes to push. Local .env file is already up to date.",
        });
      } else {
        eventBus.emit("notification", {
          type: "success",
          message: `Successfully pushed ${addedCount} new and ${changedCount} changed variable(s) to Rill Cloud.`,
        });
      }

      open = false;
      onSuccess?.();
    } catch (err) {
      // Error is already handled by the mutation
      console.error("Failed to push environment variables:", err);
    }
  }
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Push Environment Variables</DialogTitle>
      <DialogDescription>
        Push your local .env file variables to your Rill Cloud project.
      </DialogDescription>
    </DialogHeader>
    <div class="space-y-4">
      <p class="text-sm text-gray-700">
        Push your local .env file variables to your Rill Cloud project. This
        will merge your local variables with cloud variables.
      </p>

      {#if !$currentProjectQuery.data?.project}
        <div
          class="bg-yellow-50 border border-yellow-200 rounded-md p-3 text-sm text-yellow-800"
        >
          <p class="font-medium mb-1">Note</p>
          <p>
            Project will be inferred from your Git remote. Make sure your
            project is deployed to Rill Cloud.
          </p>
        </div>
      {/if}

      {#if error}
        <div
          class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-800"
        >
          <p class="font-medium mb-1">Error</p>
          <p>{error?.message || "Failed to push environment variables"}</p>
        </div>
      {/if}

      <div
        class="bg-blue-50 border border-blue-200 rounded-md p-3 text-sm text-blue-800"
      >
        <p class="font-medium mb-1">Note</p>
        <p>
          This will merge your local variables with cloud variables. Variables
          that exist in both will be updated with your local values.
        </p>
      </div>

      <p class="text-xs text-gray-500">
        You currently have <strong>{currentVariables.length}</strong>
        variable{currentVariables.length === 1 ? "" : "s"} in your local .env file.
      </p>

      <p class="text-xs text-gray-500">
        For more information see our{" "}
        <a
          href="https://docs.rilldata.com/manage/project-management/variables-and-credentials"
          target="_blank"
          rel="noopener noreferrer"
          class="text-blue-600 hover:underline"
        >
          environment variables documentation
        </a>
        .
      </p>
    </div>
    <DialogFooter>
      <Button type="plain" onClick={() => (open = false)} disabled={isPending}>
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={handlePush}
        disabled={isPending}
        loading={isPending}
      >
        Push to Rill Cloud
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
