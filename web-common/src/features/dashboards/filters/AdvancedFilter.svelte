<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let advancedFilter: V1Expression;

  $: filterText = convertExpressionToFilterParam(advancedFilter);
</script>

<div
  class="flex flex-none px-2 py-[3px] max-h-[26px] border bg-white border-slate-200 text-slate-800 rounded-2xl"
>
  <Tooltip.Root portal="body">
    <Tooltip.Trigger>
      <span class="font-bold mr-1">Advanced (BETA)</span>
      <span>{filterText}</span>
    </Tooltip.Trigger>
    <Tooltip.Content sideOffset={8}>
      Advanced filters are a bleeding edge feature! There may be bugs.
    </Tooltip.Content>
  </Tooltip.Root>
</div>
