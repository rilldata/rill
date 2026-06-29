<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { createAdminServiceTriggerReport } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getRuntimeServiceGetResourceQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useReport } from "../selectors";

  export let organization: string;
  export let project: string;
  export let report: string;

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const triggerReport = createAdminServiceTriggerReport();
  const reportQuery = useReport(runtimeClient, report);

  async function handleRunNow() {
    const lastExecution =
      $reportQuery.data?.resource.report.state.executionHistory[0];
    await $triggerReport.mutateAsync({
      org: organization,
      project,
      name: report,
      data: undefined,
    });

    eventBus.emit("notification", {
      message: m.report_triggered_adhoc(),
      type: "success",
    });

    // Refetch the resource query until the new report run shows up in the recent history table
    while (
      !$reportQuery.data ||
      $reportQuery.data.resource.report.state.executionHistory[0] ===
        lastExecution
    ) {
      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceGetResourceQueryKey(
          runtimeClient.instanceId,
          {
            name: { name: report, kind: ResourceKind.Report },
          },
        ),
      });
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
  }
</script>

<Tooltip distance={8}>
  <Button
    type="primary"
    onClick={handleRunNow}
    disabled={$triggerReport.isPending}
  >
    {m.report_run_now()}
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="300px">
    {m.report_run_now_tooltip()}
  </TooltipContent>
</Tooltip>
