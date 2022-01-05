<script>
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
export let active = true;
export let icon;
</script>

<div 
    class='collapsible-title align grid grid-cols-max justify-between'
    style="grid-template-columns: auto max-content;"
>
    <button 
        class="
            bg-transparent 
            grid 
            grid-flow-col 
            gap-2
            items-center
            p-0 
            pr-1 
            border-transparent 
            hover:border-slate-200"
            style="
                max-width: 100%;
                grid-template-columns: max-content {icon ? 'max-content' : ''} auto max-content
            "
        on:click={() => { active = !active; }}>
            <div style="
                transition: transform 100ms;
                transform: translateY(-1px) rotate({active ? 0 : -90}deg);
            "><CaretDownIcon size={14} />
        </div>
        {#if icon}
            <div class="text-gray-400">
                <svelte:component this={icon} size=1em />
            </div>
        {/if}
        <div class="text-ellipsis overflow-hidden whitespace-nowrap">
            <slot />
        </div>
        <!-- {#if caret === 'right'}
            <div style="
                transition: transform 100ms;
                transform: translateY(-1px) rotate({active ? 0 : -90}deg);
            "><CaretDownIcon size={14} /></div> 
        {/if} -->
    </button>

    <div class="contextual-information justify-self-stretch text-right">
        <slot name="contextual-information" />
    </div>
</div>

<style>
button {
    font-family: "MD IO 0.4";
    transform: translateX(-.25rem);
}
</style>