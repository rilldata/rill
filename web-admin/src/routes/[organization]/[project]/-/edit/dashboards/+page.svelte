<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import DashboardsPage from "@rilldata/web-common/features/preview-mode/DashboardsPage.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: branchPart = branchPathPrefix(branch);

  function getHref(name: string, isMetricsExplorer: boolean): string {
    const slug = isMetricsExplorer ? "explore" : "canvas";
    return `/${organization}/${project}${branchPart}/-/edit/${slug}/${name}`;
  }
</script>

<DashboardsPage {getHref} />
