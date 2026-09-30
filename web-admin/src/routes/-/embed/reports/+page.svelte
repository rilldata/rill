<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-admin/features/embeds/selectors.ts";

  const runtimeClient = useRuntimeClient();
  const personalReports = useClientFilteredResources(
    runtimeClient,
    ResourceKind.Report,
    (res) => !!res.report?.spec?.annotations?.admin_owner_user_email,
  );

  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<ReportsTable {organization} {project} data={$personalReports.data} />
