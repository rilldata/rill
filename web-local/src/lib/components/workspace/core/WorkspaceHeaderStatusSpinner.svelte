<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Spinner from "@rilldata/web-common/features/temp/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/lib/entity";

  export let applicationStatus = 0;
  let asTimer;
  function debounceStatus(status: EntityStatus) {
    clearTimeout(asTimer);
    asTimer = setTimeout(() => {
      applicationStatus = status;
    }, 100);
  }

  // TODO
  $: debounceStatus(applicationStatus);

  const applicationStatusTooltipMap = {
    [EntityStatus.Idle]: "Idle",
    [EntityStatus.Running]: "Running",
    [EntityStatus.Exporting]: "Exporting a model resultset",
    [EntityStatus.Importing]: "Importing a source",
    [EntityStatus.Profiling]: "Profiling",
  };

  $: applicationStatusTooltip = applicationStatusTooltipMap[applicationStatus];
</script>

<div class="mr-2">
  <div class="text-gray-400">
    <Tooltip alignment="center" distance={16} location="left">
      <Spinner size="18px" status={applicationStatus || EntityStatus.Idle} />
      <TooltipContent slot="tooltip-content"
        >{applicationStatusTooltip}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
