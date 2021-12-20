<script>
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
export let active = true;
</script>

<div class='collapsible-title align grid grid-cols-max grid-flow-col justify-between'>
    <button 
        class="bg-transparent w-max grid grid-flow-col gap-4 items-center p-0 pl-1 pr-1 border-transparent hover:border-slate-200" 
        on:click={() => { active = !active; }}>
        <slot />
        <div style="
            transition: transform 100ms;
            transform: translateY(-1.5px) rotate({active ? 0 : -90}deg);
        "><CaretDownIcon size={14} /></div>
    </button>

    <div class="contextual-information justify-self-stretch">
        <slot name="contextual-information" />
    </div>
</div>

<style>
button {
    font-family: "MD IO 0.4";
    transform: translateX(-.25rem);
}
</style>