<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import AlertPreview from "@rilldata/web-common/features/alerts/criteria-tab/AlertPreview.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import type { SuperForm } from "sveltekit-superforms/client";

  export let superFormInstance: SuperForm<AlertFormValues>;
  $: ({ form } = superFormInstance);
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description="Trigger alert when these conditions are met"
    title="Criteria"
  >
    <CriteriaGroup {superFormInstance} />
  </FormSection>
  <FormSection title="Alert Preview">
    <AlertPreview formValues={$form} />
  </FormSection>
</div>
