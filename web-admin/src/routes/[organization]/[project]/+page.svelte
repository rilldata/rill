<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import { WaitForDeployment } from "@rilldata/web-admin/features/projects/status/waitForDeployment";
  import { onMount } from "svelte";
  import DashboardsTable from "../../../features/dashboards/listing/DashboardsTable.svelte";

  $: project = $page.params.project;

  onMount(() => {
    WaitForDeployment.wait();
  });
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
