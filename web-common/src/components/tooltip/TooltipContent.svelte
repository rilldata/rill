<script lang="ts">
  /**
   * The maximum width of the tooltip content.
   * Defaults to 400px, which is the most commonly
   * used width for tooltips in the app.
   */
  export let maxWidth = "400px";

  /**
   * Visual variant of the tooltip.
   * - "default": Dark background (bg-tooltip) with inverse text
   * - "light": Light background (bg-popover) with normal text, theme-aware
   */
  export let variant: "default" | "light" = "default";

  $: className = $$props["class"] ?? "";

  $: variantClasses =
    variant === "light"
      ? "bg-popover text-popover-foreground"
      : "bg-tooltip text-fg-inverse";
</script>

<div
  aria-label="tooltip-content"
  class="tooltip-content shadow-md rounded p-2 pt-1 pb-1 {variantClasses} {className}"
  style:max-width={maxWidth}
>
  <slot />
</div>
