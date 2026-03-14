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

  <!-- Light Mode -->
  {#if spec?.light}
    <DescribeSection title="Light Mode">
      {#if spec.light.primary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Primary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {spec.light.primary}"
            />
            <span class="text-xs font-mono text-fg-primary"
              >{spec.light.primary}</span
            >
          </div>
        </div>
      {/if}
      {#if spec.light.secondary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Secondary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {spec.light.secondary}"
            />
            <span class="text-xs font-mono text-fg-primary"
              >{spec.light.secondary}</span
            >
          </div>
        </div>
      {/if}
      {#if spec.light.variables}
        {#each Object.entries(spec.light.variables) as [key, val]}
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
  {/if}

  <!-- Dark Mode -->
  {#if spec?.dark}
    <DescribeSection title="Dark Mode">
      {#if spec.dark.primary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Primary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {spec.dark.primary}"
            />
            <span class="text-xs font-mono text-fg-primary"
              >{spec.dark.primary}</span
            >
          </div>
        </div>
      {/if}
      {#if spec.dark.secondary}
        <div class="flex items-center justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs text-fg-secondary">Secondary</span>
          <div class="flex items-center gap-x-2">
            <span
              class="inline-block w-3 h-3 rounded-sm border border-border"
              style="background-color: {spec.dark.secondary}"
            />
            <span class="text-xs font-mono text-fg-primary"
              >{spec.dark.secondary}</span
            >
          </div>
        </div>
      {/if}
      {#if spec.dark.variables}
        {#each Object.entries(spec.dark.variables) as [key, val]}
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
  {/if}
</div>
