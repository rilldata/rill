<script>
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
export let active = true;
export let caret = 'right';
</script>

<div 
    class='collapsible-title align grid grid-cols-max grid-flow-col justify-between'
>
    <button 
        class="
            bg-transparent 
            w-max grid 
            grid-flow-col 
            gap-2
            items-center
            p-0 
            pr-1 
            border-transparent 
            hover:border-slate-200" 
        on:click={() => { active = !active; }}>
        {#if caret === 'left'}
            <div style="
                transition: transform 100ms;
                transform: translateY(-1px) rotate({active ? 0 : -90}deg);
            "><CaretDownIcon size={14} /></div>
        {/if}
        <div class="text-ellipsis">
            <slot />
        </div>
        {#if caret === 'right'}
            <div style="
                transition: transform 100ms;
                transform: translateY(-1px) rotate({active ? 0 : -90}deg);
            "><CaretDownIcon size={14} /></div>
        {/if}
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