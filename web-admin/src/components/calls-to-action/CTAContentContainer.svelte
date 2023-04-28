<script lang="ts">
  export let maxWidth = "600px";
</script>

<div
  class="flex flex-col justify-center items-center gap-y-6 w-full max-w-[{maxWidth}] text-center"
>
  <slot />
</div>
