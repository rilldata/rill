<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-admin/features/embeds/selectors.ts";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";

  const runtimeClient = useRuntimeClient();
  const personalReports = useClientFilteredResources(
    runtimeClient,
    ResourceKind.Report,
    // The runtime only returns the reports the embed user created (and all reports to admins).
    (res) => !!res.report?.spec?.annotations?.admin_embed_owner_id,
  );
  let reports = $derived($personalReports.data ?? []);

  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<ProjectPage query={personalReports} kind="report">
  <ReportsTable {organization} {project} data={reports} slot="table" />
</ProjectPage>
