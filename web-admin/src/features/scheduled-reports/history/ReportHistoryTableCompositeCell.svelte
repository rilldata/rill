<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatRunDate } from "../tableUtils";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let reportTime: string;
  export let timeZone: string;
  export let adhoc: boolean;
  export let errorMessage: string;
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-fg-primary text-sm">
    {formatRunDate(reportTime, timeZone)}
  </div>
  {#if errorMessage === ""}
    <Tag color="blue">{m.report_status_sent()}</Tag>
  {:else}
    <Tooltip distance={8}>
      <Tag color="red">{m.report_status_failed()}</Tag>
      <TooltipContent slot="tooltip-content">
        {errorMessage}
      </TooltipContent>
    </Tooltip>
  {/if}
  {#if adhoc}
    <Tooltip distance={8}>
      <Tag>{m.report_adhoc()}</Tag>
      <TooltipContent slot="tooltip-content">
        {m.report_adhoc_tooltip()}
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
