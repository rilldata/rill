<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { AlertTriangleIcon } from "lucide-svelte";
  import { DateTime } from "luxon";
  import type { CoverageWarning } from "./rollup-coverage";

  export let warning: CoverageWarning;
  export let compact = true;

  const formatDate = (ms: number) =>
    DateTime.fromMillis(ms).toLocaleString(DateTime.DATE_MED);

  $: detail = m.dashboard_partial_data_detail({
    servedStart: formatDate(warning.servedStartMs),
    availableStart: formatDate(warning.availableStartMs),
  });
</script>

<Tooltip.Root>
  <Tooltip.Trigger>
    {#snippet child({ props })}
      <span
        {...props}
        class="inline-flex items-center gap-1 text-fg-secondary text-xs"
        aria-label={m.dashboard_partial_data_hover()}
      >
        <AlertTriangleIcon class="text-amber-500" size="14px" />
        {#if !compact}
          <span>{m.dashboard_partial_data()}</span>
        {/if}
      </span>
    {/snippet}
  </Tooltip.Trigger>

  <Tooltip.Content side="top" sideOffset={6} class="max-w-xs">
    <div class="text-fg-inverse whitespace-pre-wrap break-words text-xs p-1">
      {detail}
    </div>
  </Tooltip.Content>
</Tooltip.Root>
