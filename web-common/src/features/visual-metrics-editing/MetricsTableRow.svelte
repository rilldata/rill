<script lang="ts">
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import {
    ChipColors,
    measureChipColors as colors,
  } from "@rilldata/web-common/components/chip/chip-types";

  export let measure: MetricsViewSpecMeasureV2;
</script>

<tr>
  <td>
    <input type="checkbox" />
  </td>
  <td>
    <Chip
      {...colors}
      extraRounded={false}
      extraPadding={false}
      label={measure.name}
      outline
    >
      <span slot="body" class="font-bold truncate">{measure.label}</span>
    </Chip>
  </td>
  <td class="expression">{measure.expression}</td>
  <td class="capitalize">{measure.formatPreset || measure.formatD3 || "-"}</td>
  <td>{measure.description || "-"}</td>
</tr>

<style lang="postcss">
  tr {
    @apply border-b;
    height: 40px;
  }

  td {
    @apply pl-4;
  }

  .expression {
    font-family: "Source Code Variable", monospace;
    text-transform: uppercase;
  }
</style>
