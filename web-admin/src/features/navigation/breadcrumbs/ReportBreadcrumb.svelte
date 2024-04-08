<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useReports } from "../../scheduled-reports/selectors";
  import { isReportPage } from "../nav-utils";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  $: instanceId = $runtime?.instanceId;

  $: reportName = $page.params.report;
  $: reports = useReports(instanceId);
  $: onReportPage = isReportPage($page);
</script>

{#if reportName}
  <span class="text-gray-600">/</span>
  <BreadcrumbItem
    label={reportName}
    href={`/${orgName}/${projectName}/-/reports/${reportName}`}
    menuOptions={$reports.data?.resources.map((resource) => ({
      key: resource.meta.name.name,
      main: resource.report.spec.title || resource.meta.name.name,
    }))}
    menuKey={reportName}
    onSelectMenuOption={(report) =>
      goto(`/${orgName}/${projectName}/-/reports/${report}`)}
    isCurrentPage={onReportPage}
  />
{/if}
