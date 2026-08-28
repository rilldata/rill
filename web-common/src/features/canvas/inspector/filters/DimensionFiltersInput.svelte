<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import VerticalExpressionFilters from "@rilldata/web-common/features/dashboards/filters/VerticalExpressionFilters.svelte";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";

  let {
    id,
    localExpressionFilters,
    excludedDimensions,
    updateLocalFilterString,
  }: {
    id: string;
    localExpressionFilters: ExpressionFilterManager;
    excludedDimensions: Record<string, boolean>;
    updateLocalFilterString: (newFilterString: string) => void;
  } = $props();
  // svelte-ignore state_referenced_locally
  localExpressionFilters.createListener();
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    localExpressionFilters,
    async (newUrlParams) => localExpressionFilters.setUrlParams(newUrlParams),
    () => localExpressionFilters.metricsViewsProvider.ready,
  );

  let localFiltersEnabledOverride = $state(false);

  let localFiltersEnabled = $derived(
    localExpressionFilters.hasSomeFilter || localFiltersEnabledOverride,
  );
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label={m.canvas_local_filters()}
      {id}
      faint={!localFiltersEnabled}
    />
    <Switch
      checked={localFiltersEnabled}
      onCheckedChange={() => {
        if (localFiltersEnabled) {
          localFiltersEnabledOverride = false;
          updateLocalFilterString("");
        } else {
          localFiltersEnabledOverride = true;
        }
      }}
      small
    />
  </div>
  <div class="text-fg-secondary">
    {#if localFiltersEnabled}
      {m.canvas_overriding_inherited_filters()}
    {:else}
      {m.canvas_override_inherited_filters_hint()}
    {/if}
  </div>
  {#if localFiltersEnabled}
    <VerticalExpressionFilters
      {id}
      expressionFilterManager={localExpressionFilters}
      {excludedDimensions}
    />
  {/if}
</div>
