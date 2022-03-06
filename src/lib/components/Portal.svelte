<script>
// eslint-disable-next-line import/no-extraneous-dependencies
import { onMount, onDestroy } from "svelte";
import { browser } from "$app/env"
let ref;
let portal;
let mounted = false;
onMount(() => {
    if (browser) {
        portal = document.createElement("div");
        portal.className = "portal";
        document.body.appendChild(portal);
        portal.appendChild(ref);
    }
    mounted = true;
});

onDestroy(() => {
    if (browser) {
    document.body.removeChild(portal);
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