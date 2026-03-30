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
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServicePushEnvMutation,
  } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  export let open = false;
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

  const client = useRuntimeClient();
  $: instanceQuery = createRuntimeServiceGetInstance(client, {});
  $: environment = $instanceQuery.data?.instance?.environment ?? "";

  const pushEnvMutation = createRuntimeServicePushEnvMutation(client);

  $: isPending = $pushEnvMutation.isPending;
  $: error = $pushEnvMutation.error;

  async function handlePush() {
    try {
      const result = await $pushEnvMutation.mutateAsync({});

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
  <DialogTrigger>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Push Environment Variables</DialogTitle>
      <DialogDescription>
        Merge your local .env files with cloud for {environment || "all"} environment{environment ===
        ""
          ? "s"
          : ""}. Shared keys will be updated with your local values; cloud-only
        variables are preserved.
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
        <p>{error instanceof Error ? error.message : "Failed to push environment variables"}</p>
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
