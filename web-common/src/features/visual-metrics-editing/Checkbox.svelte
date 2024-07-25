<script lang="ts">
  export let checked = false;
  export let disabled = false;
  export let onChange = () => {};
  export let size: number = 16;
</script>

<label class="form-control" style:--height="{size}px">
  <input type="checkbox" {checked} {disabled} on:change={onChange} />
</label>

<style lang="postcss">
  :root {
    --form-control-disabled: #959495;
  }

  .form-control {
    font-family: system-ui, sans-serif;
    font-size: 2rem;
    font-weight: bold;
    line-height: 1.1;
    display: grid;
    grid-template-columns: 1em auto;
    gap: 0.5em;
  }

  input[type="checkbox"] {
    -webkit-appearance: none;
    appearance: none;
    margin: 0;
    font: inherit;

    width: var(--height);
    height: var(--height);
    /* border: 0.15em solid currentColor; */
    /* border-radius: 0.15em; */
    @apply rounded-sm border border-gray-300;
    @apply bg-gray-50;
    /* transform: translateY(-0.075em); */

    display: grid;
    place-content: center;
    cursor: pointer;
  }

  input[type="checkbox"]:hover {
    @apply bg-gray-200;
  }

  input[type="checkbox"]::before {
    content: "";
    width: calc(var(--height) * 0.7);
    height: calc(var(--height) * 0.7);
    clip-path: polygon(14% 44%, 0 65%, 50% 100%, 100% 16%, 80% 0%, 43% 62%);
    transform: scale(0);
    transform-origin: bottom left;
    transition: 120ms transform ease-in-out;
    box-shadow: inset 1em 1em rgb(133, 126, 194);
    /* Windows High Contrast Mode */
    background-color: CanvasText;
  }

  input[type="checkbox"]:checked::before {
    transform: scale(1);
  }

  input[type="checkbox"]:focus {
    outline: max(2px, 0.15em) solid currentColor;
    outline-offset: max(2px, 0.15em);
  }

  input[type="checkbox"]:disabled {
    --form-control-color: var(--form-control-disabled);

    color: var(--form-control-disabled);
    cursor: not-allowed;
  }
</style>
