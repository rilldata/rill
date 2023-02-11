<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "../../features/entity-management/Spinner.svelte";

  export let applicationStatus = 0;

  let asTimer;
  function debounceStatus(status: EntityStatus) {
    clearTimeout(asTimer);
    asTimer = setTimeout(() => {
      applicationStatus = status;
    }, 500);
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

<div>
  <div class="text-gray-400">
    <Tooltip alignment="center" distance={8} location="bottom">
      <Spinner size="18px" status={applicationStatus || EntityStatus.Idle} />
      <TooltipContent slot="tooltip-content"
        >{applicationStatusTooltip}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
