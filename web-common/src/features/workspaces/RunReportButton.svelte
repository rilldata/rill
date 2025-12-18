<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceTriggerReport,
  } from "@rilldata/web-common/runtime-client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { PlayIcon } from "lucide-svelte";

  export let reportName: string;
  export let instanceId: string;

  const triggerReport = createRuntimeServiceTriggerReport(
    {
      mutation: {
        onSuccess: () => {
          eventBus.emit("notification", {
            message: `Report "${reportName}" triggered successfully`,
            type: "success",
          });
        },
        onError: (error) => {
          eventBus.emit("notification", {
            message: `Failed to trigger report: ${error.message}`,
            type: "error",
          });
        },
      },
    },
    queryClient,
  );

  async function handleTrigger() {
    await $triggerReport.mutateAsync({
      instanceId,
      name: reportName,
      data: {},
    });
  }
</script>

<Button
  type="primary"
  small
  onClick={handleTrigger}
  disabled={$triggerReport.isPending || !reportName}
  loading={$triggerReport.isPending}
>
  <PlayIcon size="15px" />
  <span>Run Now</span>
</Button>

