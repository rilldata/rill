<script lang="ts">
  import { goto } from "$app/navigation";
  import type { OlapDriver } from "@rilldata/web-common/features/connectors/olap/olap-config";
  import ProgressDots from "@rilldata/web-common/features/welcome/wizard/ProgressDots.svelte";
  import SelectFirstConnectors from "@rilldata/web-common/features/welcome/wizard/step1/SelectFirstConnectors.svelte";
  import AddFirstCredentials from "@rilldata/web-common/features/welcome/wizard/step2/AddFirstCredentials.svelte";
  import SelectTableForFirstDashboard from "@rilldata/web-common/features/welcome/wizard/step3/SelectTableForFirstDashboard.svelte";

  // Initialize variables from sessionStorage with fallbacks
  let step: 1 | 2 | 3 =
    (Number(sessionStorage.getItem("welcomeStep")) as 1 | 2 | 3) || 1;
  let managementType: "rill-managed" | "self-managed" =
    (sessionStorage.getItem("welcomeManagementType") as
      | "rill-managed"
      | "self-managed") || "rill-managed";
  let olapDriver: OlapDriver =
    (sessionStorage.getItem("welcomeOlapDriver") as OlapDriver) || "duckdb";
  let firstDataSource: string | undefined =
    sessionStorage.getItem("welcomeFirstDataSource") || undefined;

  function onSelectManagementType(type: "rill-managed" | "self-managed") {
    managementType = type;
    sessionStorage.setItem("welcomeManagementType", type);

    if (type === "rill-managed") {
      onSelectOLAP("duckdb");
    } else {
      onSelectOLAP("clickhouse");
    }

    // Reset the first data source
    firstDataSource = undefined;
    sessionStorage.removeItem("welcomeFirstDataSource");
  }

  function onSelectOLAP(olap: OlapDriver) {
    olapDriver = olap;
    sessionStorage.setItem("welcomeOlapDriver", olap);

    // reset the first data source
    firstDataSource = undefined;
    sessionStorage.removeItem("welcomeFirstDataSource");
  }

  function onSelectFirstDataSource(dataSource: string) {
    firstDataSource = dataSource;
    sessionStorage.setItem("welcomeFirstDataSource", dataSource);
  }

  function onContinueFromFirstCredentials(newFilePath: string) {
    if (firstDataSource) {
      // Go to new source file
      void goto(`/files/${newFilePath}`);
    } else {
      // Pick a table for first dashboard
      step = 3;
    }
  }

  // Update the step setter to save to sessionStorage
  $: if (step) {
    sessionStorage.setItem("welcomeStep", step.toString());
  }
</script>

<div class="container">
  {#if step === 1}
    <SelectFirstConnectors
      {managementType}
      {olapDriver}
      {firstDataSource}
      {onSelectManagementType}
      {onSelectOLAP}
      {onSelectFirstDataSource}
      onContinue={() => (step = 2)}
    />
  {:else if step === 2}
    <AddFirstCredentials
      {managementType}
      {olapDriver}
      {firstDataSource}
      onBack={() => (step = 1)}
      onContinue={onContinueFromFirstCredentials}
    />
  {:else if step === 3}
    <SelectTableForFirstDashboard />
  {:else if step === 4}
    <!-- TODO -->
  {/if}

  <span class="absolute bottom-10">
    <ProgressDots
      numberOfDots={managementType === "rill-managed" ? 2 : 3}
      activeDotIndex={step - 1}
    />
  </span>
</div>

<style lang="postcss">
  .container {
    @apply max-w-screen-xl mx-auto h-screen;
    @apply mt-12 px-4 pt-6 pb-4;
    @apply flex flex-col gap-y-4 justify-start items-center text-center;
    @apply relative;
  }
</style>
