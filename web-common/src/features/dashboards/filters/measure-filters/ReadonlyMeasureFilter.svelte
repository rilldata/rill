<script lang="ts">
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
  import { Chip } from "../../../../components/chip";
  import MeasureFilterBody from "./MeasureFilterBody.svelte";
  import type { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  let {
    measureManager,
    yamlConfigProvider,
    showPinned,
  }: {
    measureManager: MeasureFilterManager;
    yamlConfigProvider: YAMLConfigProvider;
    showPinned: boolean;
  } = $props();

  let pinned = $derived(
    Boolean(
      showPinned && yamlConfigProvider.pinnedFilters[measureManager.name],
    ),
  );
  let required = $derived(
    Boolean(
      showPinned && yamlConfigProvider.requiredFilters[measureManager.name],
    ),
  );
  let missingRequired = $derived(required && !measureManager.expr);

  let filter = $derived(<MeasureFilterEntry>{
    measure: measureManager.name,
    operation: measureManager.operation,
    type: measureManager.type,
    value1: measureManager.value1,
    value2: measureManager.value2,
  });
</script>

<Chip
  type="measure"
  theme
  label={measureManager.label}
  readOnly
  showPinnedIcon={pinned}
  error={missingRequired}
>
  <MeasureFilterBody
    slot="body"
    dimensionName={measureManager.dimension}
    {filter}
    label={measureManager.label}
  />
</Chip>
