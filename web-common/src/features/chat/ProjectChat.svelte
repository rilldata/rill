<script lang="ts">
  import FullPageChat from "./layouts/fullpage/FullPageChat.svelte";
</script>

<FullPageChat>
  <svelte:fragment slot="sidebar-footer">
    <slot name="sidebar-footer" />
  </svelte:fragment>
  <svelte:fragment slot="sidebar-collapsed-footer">
    <slot name="sidebar-collapsed-footer" />
  </svelte:fragment>
</FullPageChat>
