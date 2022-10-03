<script lang="ts">
  export let maxWidth: string = undefined;
</script>

<div
  class="bg-gray-700 text-white rounded p-2 pt-1 pb-1"
  style:max-width={maxWidth}
>
  <slot />
</div>
