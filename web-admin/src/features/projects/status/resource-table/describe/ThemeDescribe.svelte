<script lang="ts">
  import type { V1Theme } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";

  export let theme: V1Theme;

  $: spec = theme?.spec;

  function colorToHex(
    color: { red?: number; green?: number; blue?: number } | undefined,
  ): string {
    if (!color) return "";
    const r = Math.round((color.red ?? 0) * 255);
    const g = Math.round((color.green ?? 0) * 255);
    const b = Math.round((color.blue ?? 0) * 255);
    return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
  }

  $: primaryHex = spec?.primaryColorRaw || colorToHex(spec?.primaryColor);
  $: secondaryHex = spec?.secondaryColorRaw || colorToHex(spec?.secondaryColor);

  // Build mode sections data-driven to avoid duplication
  $: modeSections = [
    { label: "Light Mode", mode: spec?.light },
    { label: "Dark Mode", mode: spec?.dark },
  ].filter((s) => s.mode);
</script>

<div class="flex flex-col gap-y-3">
  <!-- Colors -->
  <DescribeSection title="Colors">
    {#if primaryHex}
      <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
        <span class="text-xs text-fg-secondary">Primary</span>
        <div class="flex items-center gap-x-2">
          <span
            class="inline-block w-3 h-3 rounded-sm border border-border"
            style="background-color: {primaryHex}"
          />
          <span class="text-xs font-mono text-fg-primary">{primaryHex}</span>
        </div>
      </div>
    {/if}
    {#if secondaryHex}
      <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
        <span class="text-xs text-fg-secondary">Secondary</span>
        <div class="flex items-center gap-x-2">
          <span
            class="inline-block w-3 h-3 rounded-sm border border-border"
            style="background-color: {secondaryHex}"
          />
          <span class="text-xs font-mono text-fg-primary">{secondaryHex}</span>
        </div>
      </div>
    {/if}
    {#if !primaryHex && !secondaryHex}
      <span class="text-xs text-fg-muted">No custom colors defined</span>
    {/if}
  </DescribeSection>

  <!-- Light / Dark Mode sections -->
  {#each modeSections as { label, mode } (label)}
    <DescribeSection title={label}>
      {#if mode.primary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Primary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {mode.primary}"
            />
            <span class="text-xs font-mono text-fg-primary">{mode.primary}</span
            >
          </div>
        </div>
      {/if}
      {#if mode.secondary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Secondary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {mode.secondary}"
            />
            <span class="text-xs font-mono text-fg-primary"
              >{mode.secondary}</span
            >
          </div>
        </div>
      {/if}
      {#if mode.variables}
        {#each Object.entries(mode.variables) as [key, val] (key)}
          <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
            <span class="text-xs text-fg-secondary">{key}</span>
            <div class="flex items-center gap-x-2">
              <span
                class="inline-block w-3 h-3 rounded-sm border border-border"
                style="background-color: {val}"
              />
              <span class="text-xs font-mono text-fg-primary">{val}</span>
            </div>
          </div>
        {/each}
      {/if}
    </DescribeSection>
  {/each}
</div>
