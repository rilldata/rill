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
    createRuntimeServicePullEnvMutation,
  } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  export let open = false;
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

  const client = useRuntimeClient();
  $: instanceQuery = createRuntimeServiceGetInstance(client, {});
  $: environment = $instanceQuery.data?.instance?.environment ?? "";

  const pullEnvMutation = createRuntimeServicePullEnvMutation(client);

  $: isPending = $pullEnvMutation.isPending;
  $: error = $pullEnvMutation.error;

  async function handlePull() {
    try {
      const result = await $pullEnvMutation.mutateAsync({});

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
  <DialogTrigger>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Pull Environment Variables</DialogTitle>
      <DialogDescription>
        Merge cloud variables into your local .env files for {environment ||
          "all"} environment{environment === "" ? "s" : ""}. Shared keys will be
        overwritten; local-only variables are preserved.
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
        <p>{error instanceof Error ? error.message : "Failed to pull environment variables"}</p>
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
        onClick={handlePull}
        disabled={isPending || !isProjectLinked}
        loading={isPending}
      >
        Pull from Rill Cloud
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
