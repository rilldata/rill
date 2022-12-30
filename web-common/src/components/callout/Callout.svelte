<script lang="ts">
  import { level as levelTokens } from "../../application-config";
  import AlertCircle from "../icons/AlertCircle.svelte";
  import AlertTriangle from "../icons/AlertTriangle.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";

  export let level: "info" | "warning" | "error" = "info";
  export let rounded = true;
  export let border = true;

  const sharedClasses = "p-3 grid gap-x-4 items-center border-1";
</script>

<div
  class:rounded
  class:border
  class="grid {sharedClasses} {levelTokens[level].text} {levelTokens[level]
    .bg} {levelTokens[level].border}"
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
