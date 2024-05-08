<script lang="ts">
  /**
   * The maximum width of the tooltip content.
   * Defaults to 400px, which is the most commonly
   * used width for tooltips in the app.
   */
  export let maxWidth = "400px";

  $: className = $$props["class"] ?? "";
</script>

<div
  aria-label="tooltip-content"
  class="tooltip-content bg-gray-700 text-white rounded p-2 pt-1 pb-1 {className}"
  style:max-width={maxWidth}
>
  <slot />
</div>
