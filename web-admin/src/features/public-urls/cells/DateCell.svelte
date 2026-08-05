<script lang="ts">
  import { getLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";

  let { value }: { value: string | undefined } = $props();

  const formatted = $derived.by(() => {
    if (!value) return "—";
    return new Date(value).toLocaleDateString(getLocale(), {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  });
</script>

<span class="whitespace-nowrap">{formatted}</span>
