<script lang="ts">
  import { browser } from "$app/environment";
  import { onDestroy, onMount } from "svelte";

  export let target: string = undefined;

  let ref;
  let portal;
  let mounted = false;
  let targetEntity;

  onMount(() => {
    if (browser) {
      portal = document.createElement("div");
      portal.className = "portal";
      if (!target) {
        targetEntity = document.body;
      } else {
        targetEntity = document.querySelector(target);
      }
      targetEntity.appendChild(portal);
      portal.appendChild(ref);
    }
    mounted = true;
  });

  onDestroy(() => {
    if (browser && portal) {
      portal.remove();
    }
  });
</script>

<div class="gp-portal" style="display: none">
  <div bind:this={ref}>
    {#if mounted}
      <slot />
    {/if}
  </div>
</div>
