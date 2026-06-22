<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let advancedFilter: V1Expression;

  $: filterText = convertExpressionToFilterParam(advancedFilter);
</script>

<div
  class="flex flex-none px-2 py-[3px] max-h-[26px] border bg-surface-subtle border-gray-200 text-fg-primary rounded-2xl"
>
  <Tooltip.Root>
    <Tooltip.Trigger>
      <span class="font-bold mr-1">{m.dashboards_filters_advanced_label()}</span
      >
      <span>{filterText}</span>
    </Tooltip.Trigger>
    <Tooltip.Content sideOffset={8}>
      {m.dashboards_filters_advanced_tooltip()}
    </Tooltip.Content>
  </Tooltip.Root>
</div>
