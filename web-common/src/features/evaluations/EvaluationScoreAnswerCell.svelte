<script lang="ts">
  import {
    formatBigNumberPercentage,
    formatDataType,
  } from "@rilldata/web-common/lib/formatters.ts";

  let {
    value,
  }: { value: { score: string; confidence: number } | undefined | null } =
    $props();

  let formattedValue = $derived(
    value
      ? `${formatDataType(value.score, "CODE_FLOAT64")} (${formatBigNumberPercentage(value.confidence)})`
      : "-",
  );
</script>

<span class="truncate">{formattedValue}</span>
