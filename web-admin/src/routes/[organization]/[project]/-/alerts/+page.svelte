<script lang="ts">
  import { page } from "$app/stores";
  import AlertsTable from "@rilldata/web-admin/features/alerts/listing/AlertsTable.svelte";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";

  const instanceId = httpClient.getInstanceId();

  $: ({
    params: { organization, project },
  } = $page);

  $: query = useAlerts(instanceId);

  $: ({ data } = $query);

  $: alerts = data?.resources ?? [];
</script>

<ProjectPage {query} kind="alert">
  <AlertsTable {organization} {project} data={alerts} slot="table" />
</ProjectPage>
