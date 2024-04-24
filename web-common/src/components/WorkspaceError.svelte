<script lang="ts">
  import AlertCircleOutline from "./icons/AlertCircleOutline.svelte";

  export let message: string;
</script>

<div class="size-full grid place-content-center">
  <div class="flex flex-col items-center gap-y-2">
    <AlertCircleOutline size="40px" />
    <h1 class="text-center">
      {message}
    </h1>
  </div>
</div>
