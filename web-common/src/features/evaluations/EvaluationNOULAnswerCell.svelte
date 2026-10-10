<script lang="ts">
  import { formatDataType } from "@rilldata/web-common/lib/formatters.ts";

  let { value }: { value: { noul: number } | undefined | null } = $props();

  let formattedValue = $derived(
    value ? formatDataType(value.noul, "CODE_FLOAT64") : "-",
  );
</script>

<span class="truncate">{formattedValue}</span>
