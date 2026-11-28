<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import { AIConditionType } from "@rilldata/web-common/features/alerts/ai-alert-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import AlertPreview from "@rilldata/web-common/features/alerts/criteria-tab/AlertPreview.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import type { SuperForm } from "sveltekit-superforms/client";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { TimeFilterManager } from "../../dashboards/time-controls/TimeFilterManager.svelte.ts";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let expressionFilterManager: ExpressionFilterManager;
  export let timeFilterManager: TimeFilterManager;

  $: ({ form, errors } = superFormInstance);
</script>

<div class="flex flex-col gap-y-3">
  {#if $form["conditionType"] === AIConditionType}
    <FormSection
      description={m.alert_form_ai_condition_description()}
      title={m.alert_form_ai_condition_title()}
    >
      <Textarea
        bind:value={$form["aiCondition"]}
        id="aiCondition"
        placeholder={m.alert_form_ai_condition_placeholder()}
        rows={4}
        errors={$errors["aiCondition"]}
      />
    </FormSection>
  {:else}
    <FormSection
      description={m.alert_form_criteria_description()}
      title={m.alert_form_criteria_title()}
    >
      <CriteriaGroup {superFormInstance} {timeFilterManager} />
    </FormSection>
    <FormSection title={m.alert_form_criteria_preview_title()}>
      <AlertPreview
        formValues={$form}
        {expressionFilterManager}
        {timeFilterManager}
      />
    </FormSection>
  {/if}
</div>
