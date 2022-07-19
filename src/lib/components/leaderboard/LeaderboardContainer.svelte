<script lang="ts">
  /** the leaderboard should have a fixed width. This enables us to
   * position the leaderboards together.
   */
  /** focused will trigger an emphasis on the leaderboard. */
  export let focused = false;
</script>

<div style:width="315px" style:margin="8px" class:drop-shadow-lg={focused}>
  <slot />
</div>
