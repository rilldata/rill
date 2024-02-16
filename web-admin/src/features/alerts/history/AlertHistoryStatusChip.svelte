<script lang="ts">
  import {
    type V1AssertionResult,
    V1AssertionStatus,
  } from "@rilldata/web-common/runtime-client";

  export let result: V1AssertionResult;

  type StatusDisplay = {
    text: string;
    textClass: string;
    borderClass: string;
  };
  const statusDisplays: Record<V1AssertionStatus, StatusDisplay> = {
    [V1AssertionStatus.ASSERTION_STATUS_UNSPECIFIED]: {
      text: "Pending",
      textClass: "text-purple-600",
      borderClass: "bg-purple-50 border-purple-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_PASS]: {
      text: "Pass",
      textClass: "text-green-600",
      borderClass: "bg-green-50 border-green-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_ERROR]: {
      text: "Errored",
      textClass: "text-red-600",
      borderClass: "bg-red-50 border-red-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_FAIL]: {
      text: "Failed",
      textClass: "text-red-600",
      borderClass: "bg-red-50 border-red-300",
    },
  };
  $: currentStatusDisplay = statusDisplays[result.status];
</script>

<div
  class="flex space-x-1 items-center px-2 border rounded rounded-[20px] w-fit {currentStatusDisplay.borderClass}"
>
  <span class={currentStatusDisplay.textClass}>{currentStatusDisplay.text}</span
  >
</div>
