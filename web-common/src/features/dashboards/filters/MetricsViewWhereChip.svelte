<!-- @component
A non-clickable chip displaying the metrics view's spec-level `where_sql`.
The backend injects this filter into every query automatically; this chip is purely informational.
-->
<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";

  export let whereSql: string;
</script>

<Tooltip.Root>
  <Tooltip.Trigger>
    <Chip type="dimension" theme readOnly gray label="Dashboard filter">
      <span slot="body" class="flex items-center gap-x-1">
        <span class="font-bold">Filter</span>
        <span class="truncate max-w-[280px]">{whereSql}</span>
      </span>
    </Chip>
  </Tooltip.Trigger>
  <Tooltip.Content sideOffset={8}>
    <div class="max-w-[480px]">
      <div class="font-bold mb-1">Dashboard filter</div>
      <div class="whitespace-pre-wrap break-words">{whereSql}</div>
      <div class="mt-1 text-fg-muted text-xs">
        Applied automatically to every query in this dashboard.
      </div>
    </div>
  </Tooltip.Content>
</Tooltip.Root>
