<script lang="ts">
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

  export let sizeBytes: string | number | undefined;

  $: formattedSize = formatSize(sizeBytes);

  function formatSize(bytes: string | number | undefined): string {
    if (bytes === undefined || bytes === null || bytes === "-1") return "-";

    let numBytes: number;
    if (typeof bytes === "number") {
      numBytes = bytes;
    } else {
      numBytes = parseInt(bytes, 10);
    }

    if (isNaN(numBytes) || numBytes < 0) return "-";
    return formatMemorySize(numBytes);
  }
</script>

<div class="truncate text-right tabular-nums">
  <span class:text-gray-500={formattedSize === "-"}>
    {formattedSize}
  </span>
</div>
