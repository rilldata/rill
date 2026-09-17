<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { SingletonProjectParserName } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { FilterGroup } from "@rilldata/web-common/components/table-toolbar/types";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { useResources } from "../selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import { filterResources } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import { getAllTagsForResources } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { DebouncedRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  let isConfirmDialogOpen = $state(false);
  const searchTextStore = new DebouncedRuneStore(
    UrlParamsState.createStringParam("q"),
    500,
  );
  const selectedTypesStore = UrlParamsState.createStringArrayParam("type");
  const selectedStatusesStore = UrlParamsState.createStringArrayParam("status");
  const selectedTagsStore = UrlParamsState.createStringArrayParam("tags");

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: m.status_filter_error(), value: "error" },
    { label: m.status_filter_warn(), value: "warn" },
    { label: m.status_filter_ok(), value: "ok" },
  ];

  // Resource types available for filtering (excluding internal types)
  const filterableTypes = [
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
    ResourceKind.Theme,
    ResourceKind.Report,
    ResourceKind.Alert,
    ResourceKind.API,
    ResourceKind.Connector,
  ];

  let resources = $derived(useResources(runtimeClient));

  // Parse errors
  let projectParserQuery = $derived(
    createRuntimeServiceGetResource(
      runtimeClient,
      {
        name: {
          kind: ResourceKind.ProjectParser,
          name: SingletonProjectParserName,
        },
      },
      { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
    ),
  );
  let parseErrors = $derived(
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [],
  );

  let hasReconcilingResources = $derived(
    $resources.data?.resources?.some(isResourceReconciling),
  );

  let isRefreshButtonDisabled = $derived(hasReconcilingResources);

  let availableTags = $derived(
    getAllTagsForResources($resources.data?.resources ?? []),
  );

  let filterGroups = $derived<FilterGroup[]>([
    {
      label: m.status_column_type(),
      key: "kind",
      options: filterableTypes.map((t) => ({
        value: t,
        label: prettyResourceKind(t),
      })),
      selectedStore: selectedTypesStore,
      defaultValue: [],
      multiSelect: true,
    },
    {
      label: m.status_label_status(),
      key: "status",
      options: statusFilters.map((s) => ({
        value: s.value,
        label: s.label,
      })),
      selectedStore: selectedStatusesStore,
      defaultValue: [],
      multiSelect: true,
    },
    ...(availableTags.length > 0
      ? [
          <FilterGroup>{
            label: m.dashboard_tags(),
            key: "tags",
            options: availableTags.map((t) => ({
              value: t.name,
              label: t.name,
            })),
            selectedStore: selectedTagsStore,
            defaultValue: [],
            multiSelect: true,
          },
        ]
      : []),
  ]);

  // Filter resources by type, search text, status, and tags
  let filteredResources = $derived(
    filterResources(
      $resources.data?.resources,
      selectedTypesStore.value,
      searchTextStore.value,
      selectedStatusesStore.value,
      selectedTagsStore.value,
    ),
  );

  function onClearAllFilters() {
    selectedTypesStore.setter([]);
    selectedStatusesStore.setter([]);
    selectedTagsStore.setter([]);
    searchTextStore.immediateSetter("");
  }

  function refreshAllSourcesAndModels() {
    void $createTrigger.mutateAsync({ all: true }).then(() => {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    });
  }
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">{m.status_nav_resources()}</h2>

  <TableToolbar {searchTextStore} {filterGroups} {onClearAllFilters}>
    <Button
      type="secondary"
      large
      class="shrink-0 whitespace-nowrap"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      <span class="hidden lg:inline">
        {m.status_refresh_all_sources_models()}
      </span>
      <span class="lg:hidden">{m.status_refresh_all()}</span>
    </Button>
  </TableToolbar>

  {#if $resources.isLoading}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      {m.status_error_loading_resources({
        error: $resources.error?.message ?? "",
      })}
    </div>
  {:else if $resources.data}
    <ProjectResourcesTable data={filteredResources} />
  {/if}

  <div class="parse-errors">
    <h3 class="parse-errors-header">
      {m.status_parse_errors_title()}
      {#if parseErrors.length > 0}
        <span class="parse-errors-badge">{parseErrors.length}</span>
      {/if}
    </h3>
    {#if parseErrors.length === 0}
      <p class="text-sm text-fg-secondary">{m.status_no_parse_errors()}</p>
    {:else}
      <div class="parse-errors-list">
        {#each parseErrors as error ((error.filePath ?? "") + ":" + error.message)}
          <div class="parse-error-item">
            {#if error.filePath}
              <span class="parse-error-file">{error.filePath}</span>
            {/if}
            <span class="parse-error-message">{error.message}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .parse-errors {
    @apply pt-4 mt-2;
  }
  .parse-errors-header {
    @apply text-sm font-semibold text-fg-primary flex items-center gap-2 mb-3;
  }
  .parse-errors-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .parse-errors-list {
    @apply flex flex-col gap-2;
  }
  .parse-error-item {
    @apply flex flex-col gap-0.5 px-3 py-2 rounded-md bg-red-50 text-sm;
  }
  .parse-error-file {
    @apply font-mono text-xs text-fg-secondary;
  }
  .parse-error-message {
    @apply text-red-700;
  }
</style>
