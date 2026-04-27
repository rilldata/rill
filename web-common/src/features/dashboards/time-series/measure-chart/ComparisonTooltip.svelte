<script lang="ts">
  export let valueFormatter: (value: number | null) => string;
  export let tooltipCurrentValue: number | null;
  export let tooltipComparisonValue: number | null;
  export let showDelta: boolean;
  export let tooltipDeltaLabel: string | null;
  export let tooltipDeltaPositive: boolean = false;
  /** When true, an increase in value is rendered as the negative (red) color. */
  export let lowerIsBetter: boolean = false;
  export let x: number;
  export let y: number;

  // Sign tracks the actual value change; color tracks whether it's favorable.
  $: deltaIsFavorable = lowerIsBetter
    ? !tooltipDeltaPositive
    : tooltipDeltaPositive;
</script>

<text class="text-outline text-[12px]" {x} {y}>
  <tspan
    class="fill-theme-700 font-semibold"
    style:font-style={tooltipCurrentValue === null ? "italic" : "normal"}
  >
    {valueFormatter(tooltipCurrentValue)}
  </tspan>
  <tspan
    class="fill-fg-muted"
    style:font-style={tooltipComparisonValue === null ? "italic" : "normal"}
  >
    vs {valueFormatter(tooltipComparisonValue)}
  </tspan>
  {#if showDelta}
    <tspan class={deltaIsFavorable ? "fill-green-600" : "fill-red-600"}>
      ({tooltipDeltaPositive ? "+" : ""}{tooltipDeltaLabel})
    </tspan>
  {/if}
</text>
