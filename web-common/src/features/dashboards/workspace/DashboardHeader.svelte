<script lang="ts">
  import { featureFlags } from "../../feature-flags";
  import DashboardControls from "./DashboardControls.svelte";
  import DashboardCTAs from "./DashboardCTAs.svelte";
  import DashboardTitle from "./DashboardTitle.svelte";

  export let metricViewName: string;

  $: isRillDeveloper = $featureFlags.readOnly === false;
</script>

<section class="w-full flex flex-col" id="header">
  <!-- top row: title and call to action -->
  {#if isRillDeveloper}
    <!-- FIXME: adding an -mb-3 fixes the spacing issue incurred by changes to the header
    to accommodate the cloud dashboard. We should go back and reconcile these headers so we don't need
  to do this. -->
    <div
      class="flex items-center justify-between -mb-3 w-full pl-1 pr-4"
      style:height="var(--header-height)"
    >
      <DashboardTitle {metricViewName} />
      <DashboardCTAs {metricViewName} />
    </div>
  {/if}
  <!-- bottom row -->
  <div class="-ml-3">
    <DashboardControls {metricViewName} />
  </div>
</section>
