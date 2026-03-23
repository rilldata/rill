<script lang="ts">
  import type { V1Theme } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";

  export let theme: V1Theme;

  $: spec = theme?.spec;

  const hexPattern = /^#[0-9a-fA-F]{3,8}$/;

  function colorToHex(
    color: { red?: number; green?: number; blue?: number } | undefined,
  ): string {
    if (!color) return "";
    const r = Math.round((color.red ?? 0) * 255);
    const g = Math.round((color.green ?? 0) * 255);
    const b = Math.round((color.blue ?? 0) * 255);
    return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
  }

  function isValidColor(value: string): boolean {
    return hexPattern.test(value);
  }

  $: primaryHex = spec?.primaryColorRaw || colorToHex(spec?.primaryColor);
  $: secondaryHex = spec?.secondaryColorRaw || colorToHex(spec?.secondaryColor);

  // Collect all color rows for the Colors section
  $: colorRows = [
    { label: "Primary", value: primaryHex },
    { label: "Secondary", value: secondaryHex },
  ].filter((r) => r.value);

  // Build mode sections data-driven to avoid duplication
  $: modeSections = [
    { label: "Light Mode", mode: spec?.light },
    { label: "Dark Mode", mode: spec?.dark },
  ].filter((s) => s.mode);

  function getModeColorRows(mode: {
    primary?: string;
    secondary?: string;
    variables?: Record<string, string>;
  }) {
    const rows: { label: string; value: string }[] = [];
    if (mode.primary) rows.push({ label: "Primary", value: mode.primary });
    if (mode.secondary)
      rows.push({ label: "Secondary", value: mode.secondary });
    if (mode.variables) {
      for (const [key, val] of Object.entries(mode.variables)) {
        rows.push({ label: key, value: val });
      }
    }
    return rows;
  }
</script>

<div class="flex flex-col gap-y-3">
  <!-- Colors -->
  <DescribeSection title="Colors">
    {#if colorRows.length > 0}
      {#each colorRows as { label, value } (label)}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">{label}</span>
          <div class="flex items-center gap-x-2">
            {#if isValidColor(value)}
              <span
                class="inline-block w-3 h-3 rounded-sm border border-border"
                style="background-color: {value}"
              />
            {/if}
            <span class="text-xs font-mono text-fg-primary">{value}</span>
          </div>
        </div>
      {/each}
    {:else}
      <span class="text-xs text-fg-muted">No custom colors defined</span>
    {/if}
  </DescribeSection>

  <!-- Light / Dark Mode sections -->
  {#each modeSections as { label, mode } (label)}
    {@const rows = getModeColorRows(mode)}
    <DescribeSection title={label}>
      {#each rows as row (row.label)}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">{row.label}</span>
          <div class="flex items-center gap-x-2">
            {#if isValidColor(row.value)}
              <span
                class="inline-block w-3 h-3 rounded-sm border border-border"
                style="background-color: {row.value}"
              />
            {/if}
            <span class="text-xs font-mono text-fg-primary">{row.value}</span>
          </div>
        </div>
      {/each}
    </DescribeSection>
  {/each}
</div>
