<script lang="ts">
  import { page } from "$app/stores";
  import AlertsTable from "@rilldata/web-admin/features/alerts/listing/AlertsTable.svelte";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";

  $: ({ instanceId } = $runtime);

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: query = useAlerts(instanceId);

  $: ({ data } = $query);

  $: alerts = data?.resources ?? [];
</script>

<ProjectPage {query} kind="alert">
  <AlertsTable {organization} {project} data={alerts} slot="table" />
  <svelte:fragment slot="action">
    To create an alert, click the "Create alert" button in a dashboard.
  </svelte:fragment>
</ProjectPage>
