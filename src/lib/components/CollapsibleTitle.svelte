<script>
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
export let active = true;
</script>

<div class='collapsible-title'>
    <button on:click={() => { active = !active; }}>
        <slot />
        <div style="
            transition: transform 100ms;
            transform: rotate({active ? 0 : -90}deg);
        "><CaretDownIcon size={14} /></div>
    </button>

    <div class="contextual-information">
        <slot name="contextual-information" />
    </div>
</div>

<style>

.collapsible-title {
    display:grid;
    grid-template-columns: max-content auto;
    justify-content: space-between;
}

button {
    border: none;
    background: none;
    border: 1px solid transparent;
    font-family: "MD IO 0.4";
    padding: 0px .25rem;
    transform: translateX(-.25rem);
    display: grid;
    grid-auto-flow: column;
    width: max-content;
    grid-gap: 1rem;
    align-items: bottom;
}

.contextual-information {
    justify-self: stretch;
}

button:hover {
    border: 1px solid hsl(217,15%, 90%);
}
</style>