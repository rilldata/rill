<script lang="ts">
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import FormItemDatePicker from "../../../components/forms/FormItemDatePicker.svelte";
  import FormItemInput from "../../../components/forms/FormItemInput.svelte";
  import FormItemSelect from "../../../components/forms/FormItemSelect.svelte";

  export let open: boolean;
  export let metricViewName: string;

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      reportName: "",
      firstRunAt: new Date().toISOString().split("T")[0], // Today's date
      frequency: "Daily",
      format: "CSV",
      limit: "",
      recipients: "",
    },
    validationSchema: yup.object({
      reportName: yup.string().required("Required"),
      firstRunAt: yup.string().required("Required"),
      frequency: yup.string().required("Required"),
      format: yup.string().required("Required"),
      // limit: yup.string().required("Required"),
      // recipients: yup.string().required("Required"),
    }),
    onSubmit: async (values) => {
      console.log(`Submit form for ${metricViewName}`, values);
    },
  });
</script>

<Dialog {open} on:close>
  <svelte:fragment slot="title">Schedule report</svelte:fragment>
  <form
    on:submit|preventDefault={handleSubmit}
    id="create-scheduled-report-form"
    autocomplete="off"
    class="flex flex-col gap-y-6"
    slot="body"
  >
    <span>Email recurring exports to recipients</span>
    <FormItemInput
      bind:value={$form["reportName"]}
      error={$errors["reportName"]}
      id="reportName"
      label="Report name"
      placeholder="My report"
    />
    <!-- error={$errors["firstRunAt"]} -->
    <FormItemDatePicker
      bind:value={$form["firstRunAt"]}
      id="firstRunAt"
      label="First run at"
    />
    <FormItemSelect
      bind:value={$form["frequency"]}
      id="frequency"
      label="Frequency"
      options={["Daily", "Weekly", "Monthly"]}
    />
    <FormItemSelect
      bind:value={$form["format"]}
      id="format"
      label="Format"
      options={["CSV", "XLSX"]}
    />
    <FormItemInput
      bind:value={$form["limit"]}
      error={$errors["limit"]}
      id="limit"
      label="Limit"
      placeholder="1000 (rows)"
    />
    <FormItemInput
      bind:value={$form["recipients"]}
      error={$errors["recipients"]}
      id="recipients"
      label="Recipients"
      placeholder="Emails separated by commas"
    />
  </form>
  <svelte:fragment slot="footer">
    <div class="flex gap-x-2 mt-6">
      <div class="grow" />
      <Button type="secondary" on:click={close}>Cancel</Button>
      <Button
        type="primary"
        submitForm
        form="create-scheduled-report-form"
        disabled={$isSubmitting}
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
