<script lang="ts">
  export let className = "";
</script>

<svg viewBox="0 0 33 33" xmlns="http://www.w3.org/2000/svg" class={className}>
  <path
    class="pan-button"
    role="presentation"
    d="M24.265 16.805L11.922 27.844C11.471 28.248 10.756 27.928 10.756 27.323L10.756 6.258C10.756 5.652 11.472 5.332 11.923 5.737L24.265 16.805Z"
  />
</svg>

<style lang="postcss">
  .pan-button {
    @apply cursor-pointer fill-slate-400;
  }
  .pan-button:hover {
    @apply fill-slate-300;
  }
</style>
