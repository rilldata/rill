<script lang="ts">
  import WarningIcon from "web-common/src/components/icons/WarningIcon.svelte";
  import type { MissingRequiredFilter } from "@rilldata/web-common/features/dashboards/filters/required/required-filters.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    missingRequiredFilters,
  }: { missingRequiredFilters: MissingRequiredFilter[] } = $props();
</script>

<div class="w-full flex justify-center px-6 pt-24 pb-12">
  <div
    class="flex flex-col items-center text-center gap-y-3 px-8 py-10 rounded-lg border border-icon-muted bg-surface-subtle shadow-sm w-full max-w-lg"
    role="alert"
  >
    <WarningIcon size="32px" className="text-amber-500" />
    <h2 class="text-lg font-semibold text-fg-primary">
      {m.dashboard_required_filters_heading()}
    </h2>
    <p class="text-sm text-fg-secondary">
      {m.dashboard_required_filters_body({
        count: missingRequiredFilters.length,
      })}
    </p>
    <ul
      class="text-sm text-fg-primary flex flex-wrap justify-center gap-x-2 gap-y-1"
    >
      {#each missingRequiredFilters as missing (missing.key)}
        <li
          class="px-2 py-0.5 rounded-md bg-red-50 border border-red-200 text-red-700"
        >
          {missing.label}
        </li>
      {/each}
    </ul>
  </div>
</div>
