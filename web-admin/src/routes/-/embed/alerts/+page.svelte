<script lang="ts">
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import AlertsTable from "@rilldata/web-admin/features/alerts/listing/AlertsTable.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-admin/features/embeds/selectors.ts";

  const runtimeClient = useRuntimeClient();
  const personalAlerts = useClientFilteredResources(
    runtimeClient,
    ResourceKind.Alert,
    (res) => !!res.alert?.spec?.annotations?.admin_owner_user_email,
  );
  let alerts = $derived($personalAlerts.data ?? []);

  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<ProjectPage query={personalAlerts} kind="alert">
  <AlertsTable {organization} {project} data={alerts} slot="table" />
</ProjectPage>
