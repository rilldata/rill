<script lang="ts">
  import { formatTokenDate } from "../personal-access-tokens";

  let { value, emptyLabel }: { value: string | undefined; emptyLabel: string } =
    $props();
</script>

<span class="whitespace-nowrap text-fg-secondary">
  {value ? formatTokenDate(value) : emptyLabel}
</span>
