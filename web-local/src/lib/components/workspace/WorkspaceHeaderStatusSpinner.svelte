<script lang="ts">
  import { getContext } from "svelte";
  import type { ApplicationStore } from "../../application-state-stores/application-store";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import { EntityStatus } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Spinner from "../Spinner.svelte";

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
    [EntityStatus.Importing]: "importing a source",
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
