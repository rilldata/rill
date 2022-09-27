<script lang="ts">
  import AlertCircle from "../icons/AlertCircle.svelte";
  import AlertTriangle from "../icons/AlertTriangle.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";

  export let level: "info" | "warning" | "error" = "info";

  const sharedClasses = "p-3 grid gap-x-4 border rounded items-center border-1";
  const classes = {
    info: "bg-blue-50 text-blue-800 border-blue-600",
    warning: "bg-yellow-100 text-yellow-800 border-yellow-600",
    error: "bg-red-100 text-red-800 border-red-600",
  };
</script>

<div
  class="grid {sharedClasses} {classes[level]}"
  style:grid-template-columns="max-content auto"
>
  <div style:font-size=".875rem">
    <slot name="icon">
      {#if level === "info"}
        <InfoCircle />
      {:else if level === "warning"}
        <AlertTriangle />
      {:else if level === "error"}
        <AlertCircle />
      {/if}
    </slot>
  </div>
  <slot />
</div>
