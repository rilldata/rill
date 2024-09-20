<script lang="ts">
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    defaultChipColors,
    excludeChipColors,
  } from "../../../../components/chip/chip-types";

  export let label: string;
  export let values: string[];
  export let isInclude: boolean;

  $: effectiveLabel = isInclude ? label : `Exclude ${label}`;
  $: colors = isInclude ? defaultChipColors : excludeChipColors;
</script>

<Chip {...colors} label={effectiveLabel} outline readOnly>
  <RemovableListBody slot="body" label={effectiveLabel} {values} />
</Chip>
