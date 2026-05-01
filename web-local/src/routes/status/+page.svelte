<script lang="ts">
  import { goto } from "$app/navigation";
  import StatusOverviewPage from "@rilldata/web-common/features/preview-mode/StatusOverviewPage.svelte";
  import { createLocalServiceGetVersion } from "@rilldata/web-common/runtime-client/local-service";
  import TablesSection from "../../features/tables/TablesSection.svelte";

  $: versionQuery = createLocalServiceGetVersion();
  $: runtimeVersion = $versionQuery.data?.current ?? null;

  function onViewResources(
    statusFilter: string[] = [],
    typeFilter: string[] = [],
  ) {
    const params = new URLSearchParams();
    if (statusFilter.length > 0) params.set("status", statusFilter.join(","));
    if (typeFilter.length > 0) params.set("kind", typeFilter.join(","));
    const search = params.toString();
    void goto(`/status/resources${search ? `?${search}` : ""}`);
  }
</script>

<StatusOverviewPage {runtimeVersion} {onViewResources}>
  <TablesSection slot="extra" />
</StatusOverviewPage>
