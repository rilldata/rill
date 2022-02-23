<script>
import Menu from "$lib/components/menu/Menu.svelte";
import MenuItem from "$lib/components/menu/MenuItem.svelte";
import Divider from "$lib/components/menu/Divider.svelte";

import CircleEmpty from "$lib/components/icons/CircleEmpty.svelte";
import CheckCircle from "$lib/components/icons/CheckCircle.svelte";

let commands = [];
function add(cmd) {
    commands.unshift(cmd);
    commands = commands.slice(0,10);
}

let active = true;


</script>

<button on:click={() => { active = !active }}>
    is active? {active}
</button>

{#if active}
<Menu 
    on:escape={() => { active = false; }} 
    on:item-select={() => { active = false; }}
>
    <MenuItem on:select={() => add('delete')}>
        <svelte:component slot="icon" this={commands[0] === 'delete' ? CheckCircle : CircleEmpty} />
        delete 
        <svelte:fragment slot="right">something</svelte:fragment>
    </MenuItem>
    <MenuItem on:select={() => add('save')}>save</MenuItem>
    <Divider />
    <MenuItem on:select={() => add('sort by')}>sort by name</MenuItem>
</Menu>
{/if}

<ul>
    {#each commands as cmd}
        <li>{cmd}</li>
    {/each}
</ul>

<div>
    This is a simple mouseover element.
</div>