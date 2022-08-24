<script lang="ts">
  import { browser } from "$app/env";
  import { Button } from "$lib/components/button";
  import { Dialog } from "$lib/components/modal-new";
  import { onMount } from "svelte";
  let active = false;
  let mounted = false;

  onMount(() => {
    mounted = true;
    // setTimeout(() => (active = true));
  });
</script>

<div style:height="200vh">
  <button
    on:click={() => {
      active = !active;
    }}>activate</button
  >

  {#if active && mounted && browser}
    <Dialog
      showCancel
      on:cancel={() => (active = false)}
      on:submit={() => (active = !active)}
    >
      <svelte:fragment slot="title">Replace existing source?</svelte:fragment>
      <svelte:fragment slot="body"
        >This action will replace all existing measures and dimensions.</svelte:fragment
      >
      <svelte:fragment slot="footer">
        <div class="flex flex-row gap-x-3 justify-items-end justify-end">
          <Button on:click={() => (active = false)} type="text">cancel</Button>
          <Button type="primary">Update source</Button>
        </div>
      </svelte:fragment>
    </Dialog>
  {/if}
</div>

<div class="fixed left-8 top-16" style:width="300px" style:height="300px">
  buncha other stuff!!!
</div>
