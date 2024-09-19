<script lang="ts">
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";

  export let label: string;
  export let values: string[];
  export let isInclude: boolean;

  $: effectiveLabel = isInclude ? label : `Exclude ${label}`;
</script>

<Chip type="dimension" label={effectiveLabel} readOnly exclude={!isInclude}>
  <RemovableListBody slot="body" label={effectiveLabel} {values} />
</Chip>
