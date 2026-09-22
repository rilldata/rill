<script lang="ts">
  import { page } from "$app/state";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { getAllTagsForResources } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { project } = $derived(page.params);

  const runtimeClient = useRuntimeClient();
  const dashboards = useDashboards(runtimeClient);
  let availableTags = $derived(getAllTagsForResources($dashboards.data ?? []));
  let hasSomeTag = $derived(availableTags.length > 0);
</script>

<svelte:head>
  <title>{m.project_overview_page_title({ project })}</title>
</svelte:head>

<ContentContainer
  maxWidth={hasSomeTag ? 1000 : 800}
  title={m.project_dashboards_title()}
>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
