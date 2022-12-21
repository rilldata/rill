<script lang="ts">
  import AlertTriangle from "../icons/AlertTriangle.svelte";

  export let message: string;
</script>

<div class="w-full">
  <div
    class="mb-2 p-2 flex bg-red-100 border-red-300 border-2 rounded text-red-800"
  >
    <AlertTriangle size="16px" />
    <p class="ml-2">
      {message}
    </p>
  </div>
</div>
