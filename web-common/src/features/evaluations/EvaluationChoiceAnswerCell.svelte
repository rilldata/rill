<script lang="ts">
  import {
    formatBigNumberPercentage,
    formatDataType,
  } from "@rilldata/web-common/lib/formatters.ts";

  let {
    value,
  }: { value: { choice: string; confidence: number } | undefined | null } =
    $props();

  let formattedValue = $derived(
    value
      ? `${formatDataType(value.choice, "VARCHAR")} (${formatBigNumberPercentage(value.confidence)})`
      : "-",
  );
</script>

<span class="truncate">{formattedValue}</span>
