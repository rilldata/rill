<script lang="ts">
  import WarningIcon from "@rilldata/web-common/components/icons/WarningIcon.svelte";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    missingRequiredFilters,
    metricsViewsProvider,
  }: {
    missingRequiredFilters: string[];
    metricsViewsProvider: MetricsViewsProvider;
  } = $props();

  let missingRequireFilterSpecs = $derived(
    missingRequiredFilters
      .map((mrf) => {
        const dimension = metricsViewsProvider.dimensionSpecs[mrf]
          ? Object.values(metricsViewsProvider.dimensionSpecs[mrf])[0]
          : undefined;
        if (dimension) {
          return {
            name: dimension.name,
            label: getDimensionDisplayName(dimension),
          };
        }

        const measure = metricsViewsProvider.measureSpecs[mrf]
          ? Object.values(metricsViewsProvider.measureSpecs[mrf])[0]
          : undefined;
        if (measure) {
          return {
            name: measure.name,
            label: getMeasureDisplayName(measure),
          };
        }

        return undefined;
      })
      .filter(Boolean) as { name: string; label: string }[],
  );
</script>

<div class="w-full flex justify-center px-6 pt-24 pb-12">
  <div
    class="flex flex-col items-center text-center gap-y-3 px-8 py-10 rounded-lg border border-gray-200 bg-surface-subtle shadow-sm w-full max-w-lg"
    role="alert"
  >
    <WarningIcon size="32px" className="text-amber-500" />
    <h2 class="text-lg font-semibold text-fg-primary">
      {m.filter_required_missing_title()}
    </h2>
    <p class="text-sm text-fg-secondary">
      {m.filter_required_missing_description({
        count: missingRequiredFilters.length,
      })}
    </p>
    <ul
      class="text-sm text-fg-primary flex flex-wrap justify-center gap-x-2 gap-y-1"
    >
      {#each missingRequireFilterSpecs as missing (missing.name)}
        <li
          class="px-2 py-0.5 rounded-md bg-red-50 border border-red-200 text-red-700"
        >
          {missing.label}
        </li>
      {/each}
    </ul>
  </div>
</div>
