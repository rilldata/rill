<script lang="ts">
  import { Tag } from "@rilldata/web-common/components/tag";
  import type { Color } from "@rilldata/web-common/components/tag/Tag.svelte";
  import {
    type V1AlertExecution,
    V1AssertionStatus,
    type V1AssertionResult,
  } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let currentExecution: V1AlertExecution;
  export let result: V1AssertionResult;

  type AssertionResultDisplay = {
    text: string;
    color: Color;
  };
  const assertionResultDisplays: Record<
    V1AssertionStatus,
    AssertionResultDisplay
  > = {
    [V1AssertionStatus.ASSERTION_STATUS_PASS]: {
      text: m.alert_status_not_triggered(),
      color: "gray",
    },
    [V1AssertionStatus.ASSERTION_STATUS_FAIL]: {
      text: m.alert_status_triggered(),
      color: "blue",
    },
    [V1AssertionStatus.ASSERTION_STATUS_ERROR]: {
      text: m.alert_status_failed(),
      color: "red",
    },
    // This should never happen
    [V1AssertionStatus.ASSERTION_STATUS_UNSPECIFIED]: {
      text: m.alert_status_unknown(),
      color: "amber",
    },
  };

  $: assertionResultDisplay = assertionResultDisplays[result.status];
</script>

{#if currentExecution}
  <Tag color="green">{m.alert_status_running()}</Tag>
{:else}
  <Tag color={assertionResultDisplay.color}>
    {assertionResultDisplay.text}
  </Tag>
{/if}
