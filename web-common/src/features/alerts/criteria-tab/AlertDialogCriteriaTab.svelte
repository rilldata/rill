<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import AlertPreview from "@rilldata/web-common/features/alerts/criteria-tab/AlertPreview.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import type { SuperForm } from "sveltekit-superforms/client";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let filters: Filters;
  export let timeControls: TimeControls;

  $: ({ form } = superFormInstance);
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description={m.alert_form_criteria_description()}
    title={m.alert_form_criteria_title()}
  >
    <CriteriaGroup {superFormInstance} {timeControls} />
  </FormSection>
  <FormSection title={m.alert_form_criteria_preview_title()}>
    <AlertPreview formValues={$form} {filters} {timeControls} />
  </FormSection>
</div>
