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
    createRuntimeServiceGetInstance,
    createRuntimeServicePullEnv,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  export let open = false;
  export let isProjectLinked = false;
  export let onSuccess: (() => void) | undefined = undefined;

  $: instanceQuery = createRuntimeServiceGetInstance($runtime.instanceId);
  $: environment = $instanceQuery.data?.instance?.environment ?? "";

  const pullEnvMutation = createRuntimeServicePullEnv();

  $: isPending = $pullEnvMutation.isPending;
  $: error = $pullEnvMutation.error;

  async function handlePull() {
    try {
      const result = await $pullEnvMutation.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {},
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
        Replace your local .env files with cloud variables for {environment ||
          "all"} environment{environment === "" ? "s" : ""}.
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
