<script lang="ts">
  import FullPageChat from "./layouts/fullpage/FullPageChat.svelte";

  export let readOnly = false;
  export let beforeFork: (() => Promise<void> | void) | undefined = undefined;
</script>

<FullPageChat {readOnly} {beforeFork}>
  <svelte:fragment slot="sidebar-footer">
    <slot name="sidebar-footer" />
  </svelte:fragment>
  <svelte:fragment slot="sidebar-collapsed-footer">
    <slot name="sidebar-collapsed-footer" />
  </svelte:fragment>
</FullPageChat>
