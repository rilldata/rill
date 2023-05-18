<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "../../icons/InfoCircle.svelte";
  import WarningIcon from "../../icons/WarningIcon.svelte";
  export let level: "error" | "warning" | "info" = undefined;
  export let message: string = undefined;
  export let line: number;
  export let active = false;
</script>

<Tooltip distance={8} suppress={message === undefined}>
  <div
    class="grid justify-between pr-2"
    style:grid-template-columns="[icon] 24px [line-number] auto [code-fold] 8px"
    class:bg-red-50={level === "error" && !active}
    class:bg-red-100={level === "error" && active}
    class:text-red-600={level === "error" && !active}
    class:text-red-700={level === "error" && active}
    class:bg-yellow-200={level === "warning" && !active}
    class:bg-yellow-300={level === "warning" && active}
    class:text-yellow-500={level === "warning" && !active}
    class:text-yellow-600={level === "warning" && active}
    class:bg-blue-200={level === "info"}
    class:bg-blue-300={level === "info" && !active}
    class:bg-blue-500={level === "info" && active}
    class:text-blue-600={level === "info" && !active}
  >
    <div
      style:grid-column="icon"
      style:height="17px"
      style:width="24px"
      class="grid justify-center items-center"
    >
      {#if level === "error"}
        <Cancel />
      {:else if level === "warning"}
        <WarningIcon />
      {:else if level === "info"}
        <InfoCircle />
      {/if}
    </div>
    <div class="text-right" style:grid-column="line-number">
      {line}
    </div>
    <!-- for code folding -->
    <div />
  </div>

  <TooltipContent maxWidth="200px" slot="tooltip-content"
    >{message}</TooltipContent
  >
</Tooltip>
