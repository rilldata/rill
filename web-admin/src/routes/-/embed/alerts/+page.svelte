<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import AlertsTable from "@rilldata/web-admin/features/alerts/listing/AlertsTable.svelte";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";

  const runtimeClient = useRuntimeClient();
  const personalAlerts = useClientFilteredResources(
    runtimeClient,
    ResourceKind.Alert,
    // The runtime only returns the alerts the embed user created (and all alerts to admins).
    (res) => !!res.alert?.spec?.annotations?.admin_embed_owner_id,
  );
  let alerts = $derived($personalAlerts.data ?? []);

  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<ProjectPage query={personalAlerts} kind="alert">
  <AlertsTable {organization} {project} data={alerts} slot="table" />
</ProjectPage>
