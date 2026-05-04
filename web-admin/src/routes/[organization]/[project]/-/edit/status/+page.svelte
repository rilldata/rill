<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import StatusOverviewPage from "@rilldata/web-common/features/preview-mode/StatusOverviewPage.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: basePath = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit/status`;

  function onViewResources(
    statusFilter: string[] = [],
    typeFilter: string[] = [],
  ) {
    const params = new URLSearchParams();
    if (statusFilter.length > 0) params.set("status", statusFilter.join(","));
    if (typeFilter.length > 0) params.set("kind", typeFilter.join(","));
    const search = params.toString();
    void goto(`${basePath}/resources${search ? `?${search}` : ""}`);
  }
</script>

<StatusOverviewPage {onViewResources} />
