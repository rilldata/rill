<script lang="ts">
  export let width = "400px";
</script>

<div
  class="flex flex-col justify-center items-center gap-y-6 w-full"
  style="max-width: {width}"
>
  <slot />
</div>
