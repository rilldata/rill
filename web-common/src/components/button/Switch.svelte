<script lang="ts">
  export let checked = false;
  export let id: string | undefined = undefined;
</script>

<button
  class="rounded flex gap-x-2 cursor-pointer select-none text-gray-800"
  on:click
>
  <slot name="left" />
  <input
    {checked}
    class="
      m-0
      checked:bg-gray-700
      cursor-pointer
    "
    {id}
    role="switch"
    type="checkbox"
  />
  <slot />
</button>

<style lang="postcss">
  input {
    @apply bg-gray-400;
    --width: 22px;
    --height: 12px;
    --margin: 3px;
    --transition: 150ms;

    appearance: none;
    -webkit-appearance: none;
    position: relative;
    display: inline-block;
    width: var(--width);
    height: var(--height);
    margin: var(--margin) 0;
    box-sizing: content-box;
    padding: 0;
    border: none;
    border-radius: 0.7em;
    /** REPLACE */
    transition: background-color var(--transition) ease;
    font-size: 100%;
    text-size-adjust: 100%;
    -webkit-text-size-adjust: 100%;
    user-select: none;
    outline: none;
  }
  input::before {
    content: "";
    display: flex;
    align-content: center;
    justify-content: center;
    position: absolute;
    width: calc(var(--height) - var(--margin));
    height: calc(var(--height) - var(--margin));
    left: 0;
    top: 0;
    @apply bg-surface;
    border-radius: 50%;
    transform: translate(calc(var(--margin) / 2), calc(var(--margin) / 2));
    transition: transform var(--transition) ease;
    line-height: 1;
  }
  input:active::before {
    background: rgba(255, 255, 255, 0.9);
  }

  input:checked {
    @apply bg-primary-500;
  }

  input:checked::before {
    transform: translate(
      calc(var(--width) - var(--height) + var(--margin) / 2),
      calc(var(--margin) / 2)
    );
  }
  input:indeterminate::before {
    transform: translate(
      calc(100% - var(--margin) * 3 / 2),
      calc(var(--margin) / 2)
    );
    content: "-";
  }
  input:disabled:before {
    opacity: 0.4;
  }
</style>
