<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  $: ({
    params: { project },
  } = $page);
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer maxWidth={800} title={m.project_dashboards_title()}>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
