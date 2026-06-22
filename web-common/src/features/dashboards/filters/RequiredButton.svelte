<script lang="ts">
  import * as BitsTooltip from "@rilldata/web-common/components/tooltip-v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let required: boolean;
  export let onToggleRequired: () => void;
</script>

<button
  class="h-full aspect-square flex items-center text-base font-bold leading-none"
  class:text-red-600={required}
  class:text-fg-secondary={!required}
  onclick={onToggleRequired}
  aria-label={required
    ? m.dashboards_filters_make_optional()
    : m.dashboards_filters_make_required()}
>
  <BitsTooltip.Root>
    <BitsTooltip.Trigger id="required-filter-tooltip">
      <span aria-hidden="true">*</span>
    </BitsTooltip.Trigger>
    <BitsTooltip.Content
      side="right"
      class="z-[100000] max-w-56 bg-gray-700 dark:bg-gray-900 shadow-md text-surface rounded p-2 pt-1 pb-1"
      sideOffset={8}
    >
      {m.dashboards_filters_required_tooltip()}
    </BitsTooltip.Content>
  </BitsTooltip.Root>
</button>
