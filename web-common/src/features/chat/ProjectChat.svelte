<script lang="ts">
  import FullPageChat from "./layouts/fullpage/FullPageChat.svelte";
</script>

<FullPageChat>
  <svelte:fragment slot="sidebar-footer">
    <slot name="sidebar-footer" />
  </svelte:fragment>
</FullPageChat>
