<script lang="ts">
  import { Chip } from "../../../../components/chip";
  import {
    measureChipColors,
    type ChipColors,
  } from "../../../../components/chip/chip-types";
  import type { V1Expression } from "../../../../runtime-client";
  import MeasureFilterBody from "./MeasureFilterBody.svelte";

  export let dimensionName: string;
  export let label: string | undefined = undefined;
  export let colors: ChipColors = measureChipColors;
  export let expr: V1Expression | undefined = undefined;
</script>

<Chip {...colors} extraRounded={false} {label} outline readOnly>
  <MeasureFilterBody {dimensionName} {expr} {label} readOnly slot="body" />
</Chip>
