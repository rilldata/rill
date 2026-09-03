<script lang="ts">
  /**
   * By default the name is kept on a single line and ellipsized. Set this to
   * let long values wrap across multiple lines instead, so the full text is
   * readable on hover.
   */
  export let wrap = false;
</script>

<div
  class="grid gap-x-2 pointer-events-none pt-1 pb-1 items-baseline"
  style="grid-template-columns: auto max-content"
  style:min-width="200px"
>
  <div
    class="font-bold text-fg-inverse"
    class:truncate={!wrap}
    class:break-words={wrap}
    aria-label="tooltip-name"
  >
    <slot name="name" />
  </div>
  <div
    class="text-fg-inverse justify-self-end pl-3"
    style:max-width="280px"
    aria-label="tooltip-name-description"
  >
    <slot name="description" />
  </div>
</div>
