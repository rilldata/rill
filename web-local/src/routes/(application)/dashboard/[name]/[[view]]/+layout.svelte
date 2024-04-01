<script lang="ts">
  import DashboardTitle from "@rilldata/web-common/features/dashboards/workspace/DashboardTitle.svelte";
  import DashboardCTAs from "@rilldata/web-common/features/dashboards/workspace/DashboardCTAs.svelte";
  import { page } from "$app/stores";
  import TimeControls from "@rilldata/web-common/features/dashboards/time-controls/TimeControls.svelte";
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  const { readOnly } = featureFlags;

  $: metricViewName = $page.params.name;

  $: isRillDeveloper = $readOnly === false;
  $: extraLeftPadding = !$navigationOpen;
</script>

<section
  class="flex flex-col size-full overflow-y-hidden dashboard-theme-boundary"
>
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 pl-4 slide"
    class:left-shift={extraLeftPadding}
  >
    {#if isRillDeveloper}
      <!-- FIXME: adding an -mb-3 fixes the spacing issue incurred by changes to the header 
    to accommodate the cloud dashboard. We should go back and reconcile these headers so we 
    don't need to do this. -->
      <div
        class="flex items-center justify-between -mb-3 w-full pl-1 pr-4"
        style:height="var(--header-height)"
      >
        <DashboardTitle {metricViewName} />
        <DashboardCTAs {metricViewName} />
      </div>
    {/if}

    <div class="-ml-3 px-1 pt-2 space-y-2">
      <TimeControls {metricViewName} />

      {#key metricViewName}
        <section class="flex justify-between gap-x-4">
          <Filters />
          <div class="flex flex-col justify-end">
            <TabBar />
          </div>
        </section>
      {/key}
    </div>
  </div>

  <slot />
</section>
