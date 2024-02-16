<script lang="ts">
  export let onCancel: () => void;
</script>

<div class="flex flex-col gap-y-1">
  <span> This could take up to 20 seconds. </span>
  <span>
    Prefer not to wait?
    <button on:click={onCancel} class="font-semibold text-primary-400"
      >Start simple</button
    >
  </span>
</div>
