<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import {
    shouldWaitForDeployment,
    waitForDeployment,
  } from "@rilldata/web-admin/features/projects/status/waitForDeployment";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import DashboardsTable from "../../../features/dashboards/listing/DashboardsTable.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  onMount(() => {
    if (get(shouldWaitForDeployment)) {
      shouldWaitForDeployment.set(false);
      waitForDeployment(organization, project);
    }
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
