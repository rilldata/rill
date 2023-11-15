<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import {
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import InputArray from "../../../components/forms/InputArray.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
    getUTCIANA,
  } from "../../../lib/time/timezone";
  import {
    convertToCron,
    getNextQuarterHour,
    getTimeIn24FormatFromDateTime,
    getTodaysDayOfWeek,
  } from "./time-utils";

  export let queryName: string;
  export let queryArgs: any;
  export let dashboardTimeZone: string;
  export let open: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();
  const createReport = createAdminServiceCreateReport();

  const user = createAdminServiceGetCurrentUser();
  const userLocalIANA = getLocalIANA();
  const UTCIana = getUTCIANA();

  // TODO: a better approach will be to use the queryArgs to craft the right state object
  const dashState = new URLSearchParams(window.location.search).get("state");

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      title: "",
      frequency: "Weekly",
      dayOfWeek: getTodaysDayOfWeek(),
      timeOfDay: getTimeIn24FormatFromDateTime(getNextQuarterHour()),
      timeZone: dashboardTimeZone || userLocalIANA,
      exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
      exportLimit: "",
      recipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
    },
    validationSchema: yup.object({
      title: yup.string().required("Required"),
      recipients: yup.array().of(
        yup.object().shape({
          email: yup.string().email("Invalid email"),
        })
      ),
    }),
    onSubmit: async (values) => {
      const refreshCron = convertToCron(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay
      );
      try {
        await $createReport.mutateAsync({
          organization,
          project,
          data: {
            options: {
              title: values.title,
              refreshCron: refreshCron, // for testing: "* * * * *"
              refreshTimeZone: values.timeZone,
              queryName: queryName,
              queryArgsJson: JSON.stringify(queryArgs),
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              openProjectSubpath: `/${queryArgs.metricsViewName}?state=${dashState}`,
              recipients: values.recipients.map((r) => r.email).filter(Boolean),
            },
          },
        });
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId)
        );
        dispatch("close");
        notifications.send({
          message: "Report created",
          type: "success",
        });
      } catch (e) {
        // showing error below
      }
    },
  });

  // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
  // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
  $: recipientErrors = $errors.recipients as unknown as { email: string }[];
</script>

<Dialog {open}>
  <svelte:fragment slot="title">Schedule report</svelte:fragment>
  <form
    autocomplete="off"
    class="flex flex-col gap-y-6"
    id="create-scheduled-report-form"
    on:submit|preventDefault={handleSubmit}
    slot="body"
  >
    <span>Email recurring exports to recipients.</span>
    <InputV2
      bind:value={$form["title"]}
      error={$errors["title"]}
      id="title"
      label="Report title"
      placeholder="My report"
    />
    <div class="flex gap-x-2">
      <Select
        bind:value={$form["frequency"]}
        id="frequency"
        label="Frequency"
        options={["Daily", "Weekdays", "Weekly"].map((frequency) => ({
          value: frequency,
        }))}
      />
      {#if $form["frequency"] === "Weekly"}
        <Select
          bind:value={$form["dayOfWeek"]}
          id="dayOfWeek"
          label="Day"
          options={[
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
            "Saturday",
            "Sunday",
          ].map((day) => ({
            value: day,
          }))}
        />
      {/if}
      <TimePicker bind:value={$form["timeOfDay"]} id="timeOfDay" label="Time" />
      <Select
        bind:value={$form["timeZone"]}
        id="timeZone"
        label="Time zone"
        options={[dashboardTimeZone, userLocalIANA, UTCIana]
          // Remove duplicates when dashboardTimeZone is already covered by userLocalIANA or UTCIana
          .filter((z, i, self) => {
            return self.indexOf(z) === i;
          })
          // Add labels
          .map((z) => {
            let label = getAbbreviationForIANA(new Date(), z);
            if (z === userLocalIANA) {
              label += " (local time)";
            }
            return {
              value: z,
              label: label,
            };
          })}
      />
    </div>
    <Select
      bind:value={$form["exportFormat"]}
      id="exportFormat"
      label="Format"
      options={[
        { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
        { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
        { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "Excel" },
      ]}
    />
    <InputV2
      bind:value={$form["exportLimit"]}
      error={$errors["exportLimit"]}
      id="exportLimit"
      label="Row limit"
      optional
      placeholder="1000"
    />
    <InputArray
      id="recipients"
      label="Recipients"
      bind:values={$form["recipients"]}
      bind:errors={recipientErrors}
      accessorKey="email"
      hint="Recipients will receive different views based on their security policy.
        Recipients without project access can't view the report."
      placeholder="Enter an email address"
      addItemLabel="Add email"
      on:add-item={() => {
        $form["recipients"] = $form["recipients"].concat({ email: "" });
        recipientErrors = recipientErrors.concat({ email: "" });
      }}
      on:remove-item={(event) => {
        const index = event.detail.index;
        $form["recipients"] = $form["recipients"].filter((r, i) => i !== index);
        recipientErrors = recipientErrors.filter((r, i) => i !== index);
      }}
    />
  </form>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-5">
      {#if $createReport.isError}
        <div class="text-red-500">{$createReport.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button
        disabled={$isSubmitting ||
          $form["recipients"].filter((r) => r.email).length === 0}
        form="create-scheduled-report-form"
        submitForm
        type="primary"
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
