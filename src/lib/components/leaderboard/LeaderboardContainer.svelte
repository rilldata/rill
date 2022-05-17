<script lang="ts">
  /** the leaderboard should have a fixed width. This enables us to
   * position the leaderboards together.
   */
  export let width = 315;
  /** focused will trigger an emphasis on the leaderboard. */
  export let focused = false;
</script>

<div class:drop-shadow-lg={focused} style:width="{width}px">
  <slot />
</div>
