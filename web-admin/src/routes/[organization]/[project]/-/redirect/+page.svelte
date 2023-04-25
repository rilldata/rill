<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  // Currently, we don't have an `/[organization]` page. Instead, `/[organization]` routes to the organization's first project, and ideally that project's first dashboard.
  // This page is used to redirect to either the project's first dashboard, or to the project's status page.
  $: dashboardsQuery = useDashboardNames($runtime.instanceId);
  $: if ($dashboardsQuery.isSuccess) {
    if ($dashboardsQuery.data.length === 0) {
      goto(`/${$page.params.organization}/${$page.params.project}`);
    } else {
      goto(
        `/${$page.params.organization}/${$page.params.project}/${$dashboardsQuery.data[0]}`
      );
    }
  }
</script>
