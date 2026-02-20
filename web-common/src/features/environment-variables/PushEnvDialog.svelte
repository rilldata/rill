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
  import { createRuntimeServicePushEnv } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  export let open = false;
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

  let selectedEnvironment: "dev" | "prod" | "both" = "both";

  const pushEnvMutation = createRuntimeServicePushEnv();

  $: isPending = $pushEnvMutation.isPending;
  $: error = $pushEnvMutation.error;

  async function handlePush() {
    try {
      const result = await $pushEnvMutation.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          environment:
            selectedEnvironment === "both" ? "" : selectedEnvironment,
        },
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
        Merge your local .env variables with cloud. Existing cloud variables
        will be updated with your local values.
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
        <button
          type="button"
          class="px-3 py-1.5 text-sm rounded-md border transition-colors {selectedEnvironment ===
          'both'
            ? 'bg-primary-100 border-primary-500 text-primary-600'
            : 'bg-surface border text-fg-secondary hover:bg-surface-hover'}"
          on:click={() => (selectedEnvironment = "both")}
        >
          Both
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
        <p>{error?.message || "Failed to push environment variables"}</p>
      </div>
    {/if}

    <DialogFooter>
      <Button
        type="tertiary"
        onClick={() => (open = false)}
        disabled={isPending}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={handlePush}
        disabled={isPending || !isProjectLinked}
        loading={isPending}
      >
        Push to Rill Cloud
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
