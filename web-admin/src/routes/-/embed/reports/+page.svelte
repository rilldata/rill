<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";

  const runtimeClient = useRuntimeClient();
  const personalReports = useClientFilteredResources(
    runtimeClient,
    ResourceKind.Report,
    (res) => !!res.report?.spec?.annotations?.personal,
  );
</script>

<ReportsTable organization="" project="" data={$personalReports.data} />
