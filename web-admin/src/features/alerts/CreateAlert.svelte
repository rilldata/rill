<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import {
    DialogContent,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog/index";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { BellPlusIcon } from "lucide-svelte";

  const {
    selectors: {
      timeRangeSelectors: { isCustomTimeRange },
    },
    metricsViewName,
    exploreName,
    dashboardStore,
  } = getStateManagers();

  const runtimeClient = useRuntimeClient();

  $: metricsView = useMetricsViewValidSpec(runtimeClient, $metricsViewName);
  $: hasTimeDimension = !!$metricsView?.data?.timeDimension;

  let open = false;
</script>

{#if hasTimeDimension && $dashboardStore}
  <GuardedDialog
    title={m.dialog_close_without_saving_title()}
    description={m.dialog_close_without_saving_alert_desc()}
    confirmLabel={m.dialog_close_without_saving_confirm()}
    cancelLabel={m.dialog_close_without_saving_cancel()}
    bind:open
    let:onCancel
    let:onClose
    let:preventClose
  >
    <DialogTrigger>
      {#snippet child({ props })}
        <Tooltip distance={8} location="top" suppress={!$isCustomTimeRange}>
          <Button
            {...props}
            compact
            disabled={$isCustomTimeRange}
            type="secondary"
            label={m.alert_create_alert()}
          >
            <BellPlusIcon class="inline-flex" size="16px" />
          </Button>
          <TooltipContent slot="tooltip-content">
            {m.alert_set_non_custom_time_range()}
          </TooltipContent>
        </Tooltip>
      {/snippet}
    </DialogTrigger>
    <DialogContent
      class="p-0 m-0 w-[802px] max-w-fit rounded-md"
      noClose
      onEscapeKeydown={preventClose}
      onInteractOutside={preventClose}
    >
      <AlertFormDataWrapper
        props={{ mode: "create", exploreName: $exploreName }}
        {onCancel}
        {onClose}
      />
    </DialogContent>
  </GuardedDialog>
{/if}
