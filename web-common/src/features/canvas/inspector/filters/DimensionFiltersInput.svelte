<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import VerticalExpressionFilters from "@rilldata/web-common/features/dashboards/filters/VerticalExpressionFilters.svelte";
  import { onDestroy } from "svelte";

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
  const stateChangeUnsub = localExpressionFilters.on("state-changed", () => {
    updateLocalFilterString(
      Object.values(localExpressionFilters.paramByManager)[0] ?? "",
    );
  });

  let localFiltersEnabledOverride = $state(false);

  let localFiltersEnabled = $derived(
    localExpressionFilters.hasSomeFilter || localFiltersEnabledOverride,
  );

  onDestroy(() => stateChangeUnsub());
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
