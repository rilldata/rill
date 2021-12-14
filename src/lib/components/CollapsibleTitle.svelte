<script>
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
export let active = true;
</script>

<button on:click={() => { active = !active; }}>
    <slot />
    <div style="
        transition: transform 100ms;
        transform: rotate({active ? 0 : -90}deg);
    "><CaretDownIcon size={14} /></div>
</button>

<style>
button {
    border: none;
    background: none;
    display: grid;
    grid-auto-flow: column;
    width: max-content;
    grid-gap: 1rem;
    align-items: bottom;
    border: 1px solid transparent;
    font-family: "MD IO 0.4";
}

button:hover {
    border: 1px solid gray;
}
</style>