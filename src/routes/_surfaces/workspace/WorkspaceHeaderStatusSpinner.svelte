<script lang="ts">
  import { getContext } from "svelte";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Spinner from "$lib/components/Spinner.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;

  let applicationStatus = 0;
  let asTimer;
  function debounceStatus(status: EntityStatus) {
    clearTimeout(asTimer);
    asTimer = setTimeout(() => {
      applicationStatus = status;
    }, 100);
  }

  $: debounceStatus($store?.status as unknown as EntityStatus);

  const applicationStatusTooltipMap = {
    [EntityStatus.Idle]: "idle",
    [EntityStatus.Running]: "running",
    [EntityStatus.Exporting]: "exporting a model resultset",
    [EntityStatus.Importing]: "importing a table",
    [EntityStatus.Profiling]: "profiling",
  };

  $: applicationStatusTooltip = applicationStatusTooltipMap[applicationStatus];
</script>

<div>
  <div class="text-gray-400">
    <Tooltip location="left" alignment="center" distance={16}>
      <Spinner status={applicationStatus || EntityStatus.Idle} size="20px" />
      <TooltipContent slot="tooltip-content"
        >{applicationStatusTooltip}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
