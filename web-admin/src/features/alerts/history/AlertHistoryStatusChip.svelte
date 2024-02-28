<script lang="ts">
  import {
    V1AssertionStatus,
    type V1AssertionResult,
  } from "@rilldata/web-common/runtime-client";

  export let result: V1AssertionResult;

  type StatusDisplay = {
    text: string;
    textClass: string;
    borderClass: string;
  };
  const statusDisplays: Record<V1AssertionStatus, StatusDisplay> = {
    [V1AssertionStatus.ASSERTION_STATUS_UNSPECIFIED]: {
      // This should never happen
      text: "Status unknown",
      textClass: "text-yellow-600",
      borderClass: "bg-yellow-50 border-yellow-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_PASS]: {
      text: "Not triggered",
      textClass: "text-gray-600",
      borderClass: "bg-gray-50 border-gray-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_FAIL]: {
      text: "Triggered",
      textClass: "text-blue-600",
      borderClass: "bg-blue-50 border-blue-300",
    },
    [V1AssertionStatus.ASSERTION_STATUS_ERROR]: {
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
