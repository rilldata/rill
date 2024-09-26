<script lang="ts">
  export let checked = false;
  export let disabled = false;
  export let onChange: (checked: boolean) => void;
  export let size: number = 16;
</script>

<label class="form-control" style:--height="{size}px">
  <input
    type="checkbox"
    {checked}
    {disabled}
    on:change={(event) => {
      onChange(event.currentTarget.checked);
    }}
  />
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
    width: var(--height);
    height: var(--height);
  }

  input[type="checkbox"] {
    -webkit-appearance: none;
    appearance: none;
    margin: 0;
    font: inherit;

    width: var(--height);
    height: var(--height);

    @apply rounded-[2px] border border-gray-300;
    @apply bg-gray-50;

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

    box-shadow: inset 1em 1em rgb(255, 255, 255);
    /* Windows High Contrast Mode */
    background-color: CanvasText;
  }

  input[type="checkbox"]:checked::before {
    transform: scale(1);
    background-color: rgb(35, 16, 207);
  }

  input[type="checkbox"]:checked {
    transform: scale(1);
    @apply bg-primary-400;
    @apply border-primary-400;
  }

  input[type="checkbox"]:disabled {
    --form-control-color: var(--form-control-disabled);

    color: var(--form-control-disabled);
    cursor: not-allowed;
  }
</style>
