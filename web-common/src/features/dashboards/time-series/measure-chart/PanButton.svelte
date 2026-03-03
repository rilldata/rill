<script lang="ts">
  export let onClick: () => void;
  export let flip: boolean = false;
</script>

<g
  class="pan-controls"
  transform={flip ? "translate(32, 0) scale(-1, 1)" : undefined}
>
  <path
    role="presentation"
    d="M9.335 16.795L21.678 5.756C22.129 5.352 22.844 5.672 22.844 6.277L22.844 27.342C22.844 27.948 22.128 28.268 21.677 27.863L9.335 16.795Z"
    class="pan-button"
    on:click|stopPropagation={onClick}
  />
</g>

<style lang="postcss">
  .pan-button {
    @apply cursor-pointer fill-icon-muted;
    @apply opacity-60;
  }
  .pan-button:hover {
    @apply opacity-100;
  }
</style>
