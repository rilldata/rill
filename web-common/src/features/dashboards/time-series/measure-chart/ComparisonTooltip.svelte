<script lang="ts">
  export let valueFormatter: (value: number | null) => string;
  export let tooltipCurrentValue: number | null;
  export let tooltipComparisonValue: number | null;
  export let showDelta: boolean;
  export let tooltipDeltaLabel: string | null;
  export let tooltipDeltaPositive: boolean = false;
  export let x: number;
  export let y: number;
</script>

<text class="text-outline text-[12px]" {x} {y}>
  <tspan class="fill-theme-700 font-semibold">
    {valueFormatter(tooltipCurrentValue)}
  </tspan>
  <tspan class="fill-fg-muted">
    vs {valueFormatter(tooltipComparisonValue)}
  </tspan>
  {#if showDelta}
    <tspan class={tooltipDeltaPositive ? "fill-green-600" : "fill-red-600"}>
      ({tooltipDeltaPositive ? "+" : ""}{tooltipDeltaLabel})
    </tspan>
  {/if}
</text>
