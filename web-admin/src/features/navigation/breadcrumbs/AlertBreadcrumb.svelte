<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { isAlertPage } from "../nav-utils";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  $: instanceId = $runtime?.instanceId;

  $: alertName = $page.params.alert;
  $: alerts = useAlerts(instanceId);
  $: onAlertPage = isAlertPage($page);
</script>

{#if alertName}
  <span class="text-gray-600">/</span>
  <BreadcrumbItem
    label={alertName}
    href={`/${orgName}/${projectName}/-/alerts/${alertName}`}
    menuItems={$alerts.data?.resources.map((resource) => ({
      key: resource.meta.name.name,
      main: resource.alert.spec.title || resource.meta.name.name,
    }))}
    menuKey={alertName}
    onSelectMenuItem={(alert) =>
      goto(`/${orgName}/${projectName}/-/alerts/${alert}`)}
    isCurrentPage={onAlertPage}
  />
{/if}
