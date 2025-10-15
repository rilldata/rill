<script lang="ts">
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let exploreName: string;
  export let metricsViewName: string;

  $: ({ instanceId } = $runtime);
  $: explore = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $explore.data?.explore;

  $: timeRanges = exploreSpec?.timeRanges ?? [];
  $: hasTimeSeries = !!$explore.data?.metricsView?.timeDimension;
</script>

<article
  class="flex flex-col size-full overflow-y-hidden dashboard-theme-boundary"
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-background slide"
  >
    <section class="flex relative justify-between gap-x-4 py-4 pb-6 px-4">
      <Filters {timeRanges} {metricsViewName} {hasTimeSeries} />
    </section>
  </div>
  <PivotDisplay />
</article>
