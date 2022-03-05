<script>
import PaneExpanderIcon from "$lib/components/PaneExpanderIcon.svelte"
let mode = 'hamburger;'
</script>

<button on:click={() => { mode = 'left'; }}>
    left
</button>
<button on:click={() => { mode = 'hamburger'; }}>
    hamburger
</button>

<button on:click={() => { mode = 'right'; }}>
    right
</button>

<PaneExpanderIcon size="5rem" {mode} />
<PaneExpanderIcon size="4rem"  {mode} />
<PaneExpanderIcon size="3rem" {mode}  />
<PaneExpanderIcon size="2rem"  {mode} />
<PaneExpanderIcon size="24px"  {mode} />
<PaneExpanderIcon size="1rem"  {mode} />