<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1ListResourcesResponse } from "@rilldata/web-common/runtime-client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";

  export let kind: "report" | "dashboard" | "alert";
  export let query: CreateQueryResult<V1ListResourcesResponse, Error>;

  $: ({ isLoading, isError, isSuccess, error } = $query);

  const kindTitleMap: Record<string, () => string> = {
    report: () => m.nav_tab_reports(),
    dashboard: () => m.nav_tab_dashboards(),
    alert: () => m.nav_tab_alerts(),
  };
  $: title = (kindTitleMap[kind] ?? (() => kind))();
</script>

<ContentContainer {title}>
  <div class="flex flex-col items-center gap-y-4">
    {#if isLoading}
      <div class="m-auto mt-20">
        <DelayedSpinner isLoading size="24px" />
      </div>
    {:else if isError}
      <ResourceError {kind} {error} />
    {:else if isSuccess}
      <slot name="table" />
    {/if}
  </div>
</ContentContainer>
