<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "../../../features/dashboards/listing/DashboardsTable.svelte";
  import { useDashboardsV2 } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({
    params: { project },
  } = $page);
  $: ({ instanceId } = $runtime);

  $: query = useDashboardsV2(instanceId);
  $: ({ data } = $query);
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer
  maxWidth={800}
  title="Project dashboards"
  showTitle={data?.length > 0}
>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
