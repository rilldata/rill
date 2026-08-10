<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { V1AIEvalVerdict } from "@rilldata/web-common/runtime-client";

  export let verdict: V1AIEvalVerdict | undefined;

  $: [label, style] = display(verdict);

  function display(v: V1AIEvalVerdict | undefined): [string, string] {
    switch (v) {
      case V1AIEvalVerdict.AI_EVAL_VERDICT_PASS:
        return [m.eval_verdict_pass(), "pass"];
      case V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL:
        return [m.eval_verdict_fail(), "fail"];
      case V1AIEvalVerdict.AI_EVAL_VERDICT_ERROR:
        return [m.eval_verdict_error(), "error"];
      case V1AIEvalVerdict.AI_EVAL_VERDICT_SKIPPED:
        return [m.eval_verdict_skipped(), "muted"];
      case V1AIEvalVerdict.AI_EVAL_VERDICT_RUNNING:
        return [m.eval_verdict_running(), "running"];
      case V1AIEvalVerdict.AI_EVAL_VERDICT_PENDING:
        return [m.eval_verdict_pending(), "muted"];
      default:
        return ["", "muted"];
    }
  }
</script>

{#if label}
  <span class="chip {style}">{label}</span>
{/if}

<style lang="postcss">
  .chip {
    @apply shrink-0 rounded-full px-2 py-0.5 text-xs font-medium;
  }

  .chip.pass {
    @apply bg-green-100 text-green-800;
  }

  .chip.fail {
    @apply bg-red-100 text-red-800;
  }

  .chip.error {
    @apply bg-yellow-100 text-yellow-800;
  }

  .chip.running {
    @apply bg-primary-100 text-primary-800;
  }

  .chip.muted {
    @apply bg-surface-hover text-fg-muted;
  }
</style>
