<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import { AlertCircleIcon } from "lucide-svelte";

  export let name: string;
  export let measure: string;
  export let splitByDimension: string;
  export let criteria: MeasureFilterEntry[];
  export let criteriaOperation: V1Operation;
  export let snooze: string;
  export let enableSlackNotification: boolean;
  export let slackChannels: string[];
  export let slackUsers: string[];
  export let enableEmailNotification: boolean;
  export let emailRecipients: string[];
  export let refreshWhenDataRefreshes: boolean;
  export let frequency: string;
  export let dayOfWeek: string;
  export let timeOfDay: string;
  export let dayOfMonth: string;
  export let timeZone: string;
  export let metricsViewSpec: V1MetricsViewSpec;
  export let hasTimeComparison: boolean;
  export let errors: LineStatus[] = [];

  // Extract error messages for display
  $: errorMessages = errors.map((e) => e.message).filter(Boolean);
  $: hasErrors = errors.length > 0;

  // Data source specific fields
  export let metricsSql: string = "";
  export let sql: string = "";
  export let sqlConnector: string = "";
  export let resourceError: string = "";
  export let whenErrors: boolean = false;

  // Data source type - derive from incoming props
  type DataSourceType = "metrics_sql" | "sql" | "resource_error";
  let selectedDataSource: DataSourceType = "metrics_sql";

  // Initialize selectedDataSource based on which data prop has a value
  $: {
    if (whenErrors) {
      selectedDataSource = "resource_error";
    } else if (sql) {
      selectedDataSource = "sql";
    } else if (metricsSql) {
      selectedDataSource = "metrics_sql";
    }
  }

  // Handle data source selection - clear others and set default for selected
  function selectDataSource(type: DataSourceType) {
    selectedDataSource = type;
    if (type === "metrics_sql") {
      metricsSql = metricsSql || "SELECT measure FROM metrics_view";
      sql = "";
      whenErrors = false;
    } else if (type === "sql") {
      sql = sql || "SELECT column FROM table";
      metricsSql = "";
      whenErrors = false;
    } else if (type === "resource_error") {
      metricsSql = "";
      sql = "";
      whenErrors = true; // Default to true when selecting resource status
    }
  }

  const dataSourceTypes: {
    value: DataSourceType;
    label: string;
    description: string;
  }[] = [
    {
      value: "metrics_sql",
      label: "Metrics SQL",
      description: "Query metrics views using SQL-like syntax",
    },
    {
      value: "sql",
      label: "SQL",
      description: "Query models/sources using raw SQL",
    },
    {
      value: "resource_error",
      label: "Resource Status",
      description: "Trigger based on resource status",
    },
  ];

  // Cron expression for custom schedule
  export let cron: string = "";

  // Schedule type - derive from refreshWhenDataRefreshes prop
  type ScheduleType = "data_refresh" | "cron";
  let scheduleType: ScheduleType = "data_refresh";

  // Initialize scheduleType based on refreshWhenDataRefreshes
  $: {
    if (refreshWhenDataRefreshes) {
      scheduleType = "data_refresh";
    } else if (cron) {
      scheduleType = "cron";
    }
  }

  // Handle schedule type change from user interaction
  function handleScheduleTypeChange(newType: ScheduleType) {
    scheduleType = newType;
    if (newType === "data_refresh") {
      refreshWhenDataRefreshes = true;
      cron = "";
    } else {
      refreshWhenDataRefreshes = false;
    }
  }

  // Helper to convert slack inputs to arrays
  function parseCommaSeparated(value: string): string[] {
    return value
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
  }

  // Local string values for comma-separated inputs
  let slackChannelsInput = "";
  let slackUsersInput = "";
  let emailRecipientsInput = "";
  let inputsInitialized = false;

  // Track the last known array values to detect external changes
  let lastSlackChannels: string[] = [];
  let lastSlackUsers: string[] = [];
  let lastEmailRecipients: string[] = [];

  // Initialize inputs from props and sync when props change externally
  $: {
    const slackChannelsChanged = JSON.stringify(slackChannels) !== JSON.stringify(lastSlackChannels);
    const slackUsersChanged = JSON.stringify(slackUsers) !== JSON.stringify(lastSlackUsers);
    const emailRecipientsChanged = JSON.stringify(emailRecipients) !== JSON.stringify(lastEmailRecipients);

    if (!inputsInitialized || slackChannelsChanged) {
      slackChannelsInput = slackChannels.join(", ");
      lastSlackChannels = [...slackChannels];
    }
    if (!inputsInitialized || slackUsersChanged) {
      slackUsersInput = slackUsers.join(", ");
      lastSlackUsers = [...slackUsers];
    }
    if (!inputsInitialized || emailRecipientsChanged) {
      emailRecipientsInput = emailRecipients.join(", ");
      lastEmailRecipients = [...emailRecipients];
    }
    inputsInitialized = true;
  }

  // Keep arrays in sync with inputs when user types
  function handleSlackChannelsInput(newValue: string) {
    slackChannelsInput = newValue;
    slackChannels = parseCommaSeparated(newValue);
    lastSlackChannels = [...slackChannels];
  }

  function handleSlackUsersInput(newValue: string) {
    slackUsersInput = newValue;
    slackUsers = parseCommaSeparated(newValue);
    lastSlackUsers = [...slackUsers];
  }

  function handleEmailRecipientsInput(newValue: string) {
    emailRecipientsInput = newValue;
    emailRecipients = parseCommaSeparated(newValue);
    lastEmailRecipients = [...emailRecipients];
  }

  // Suppress unused variable warnings
  void measure;
  void splitByDimension;
  void criteria;
  void criteriaOperation;
  void snooze;
  void enableSlackNotification;
  void enableEmailNotification;
  void frequency;
  void dayOfWeek;
  void timeOfDay;
  void dayOfMonth;
  void timeZone;
  void metricsViewSpec;
  void hasTimeComparison;
  void sqlConnector;
</script>

<div class="h-full w-full flex flex-col bg-surface-background overflow-hidden">
  <!-- Content -->
  <div class="flex-1 overflow-y-auto p-6">
    <div class="flex flex-col gap-y-6">
      <!-- Alert Name -->
      <div class="flex flex-col gap-y-2">
        <Input
          id="alert-name"
          label="Name"
          hint="A descriptive name for this alert"
          bind:value={name}
          placeholder="e.g. High revenue alert"
        />
      </div>

      <!-- Data Section -->
      <div class="flex flex-col gap-y-3">
        <Label class="text-sm font-semibold text-fg-primary">Data</Label>

        <!-- Card Style Type Buttons (matching API PR) -->
        <div class="grid grid-cols-3 gap-2">
          {#each dataSourceTypes as dataType}
            <button
              class="type-button"
              class:selected={selectedDataSource === dataType.value}
              on:click={() => selectDataSource(dataType.value)}
            >
              <span class="type-label">{dataType.label}</span>
              <span class="type-description">{dataType.description}</span>
            </button>
          {/each}
        </div>

        <!-- Data Source Content -->
        {#if selectedDataSource === "metrics_sql"}
          <div class="flex flex-col gap-y-3 p-4 border rounded-md bg-surface-subtle">
            <Label class="text-sm text-fg-secondary">Metrics SQL Query</Label>
            <textarea
              bind:value={metricsSql}
              class="w-full h-32 p-3 border rounded-md font-mono text-sm bg-surface-background resize-y focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="SELECT measure FROM metrics_view"
            />
          </div>
        {:else if selectedDataSource === "sql"}
          <div class="flex flex-col gap-y-3 p-4 border rounded-md bg-surface-subtle">
            <Label class="text-sm text-fg-secondary">SQL Query</Label>
            <textarea
              bind:value={sql}
              class="w-full h-32 p-3 border rounded-md font-mono text-sm bg-surface-background resize-y focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="SELECT column FROM table"
            />
          </div>
        {:else if selectedDataSource === "resource_error"}
          <div class="flex flex-col gap-y-3 p-4 border rounded-md bg-surface-subtle">
            <label class="flex items-center gap-x-2 cursor-pointer">
              <input
                type="checkbox"
                bind:checked={whenErrors}
                class="w-4 h-4 text-primary-600 rounded"
              />
              <span class="text-sm">Trigger when resource is in error state (where_error)</span>
            </label>
          </div>
        {/if}
      </div>

      <!-- Schedule Section -->
      <div class="flex flex-col gap-y-3">
        <Label class="text-sm font-semibold text-fg-primary">Schedule</Label>

        <div class="flex gap-x-4">
          <label class="flex items-center gap-x-2 cursor-pointer">
            <input
              type="radio"
              name="schedule"
              value="data_refresh"
              checked={scheduleType === "data_refresh"}
              on:change={() => handleScheduleTypeChange("data_refresh")}
              class="w-4 h-4 text-primary-600"
            />
            <span class="text-sm">When data refreshes</span>
          </label>
          <label class="flex items-center gap-x-2 cursor-pointer">
            <input
              type="radio"
              name="schedule"
              value="cron"
              checked={scheduleType === "cron"}
              on:change={() => handleScheduleTypeChange("cron")}
              class="w-4 h-4 text-primary-600"
            />
            <span class="text-sm">Custom schedule</span>
          </label>
        </div>

        {#if scheduleType === "cron"}
          <div class="p-4 border rounded-md bg-surface-subtle">
            <Input
              id="cron"
              label="Cron Expression"
              hint="e.g. 0 9 * * * (daily at 9am)"
              bind:value={cron}
              placeholder="0 9 * * *"
            />
          </div>
        {/if}
      </div>

      <!-- Notify Section -->
      <div class="flex flex-col gap-y-3">
        <Label class="text-sm font-semibold text-fg-primary">Notify</Label>

        <div class="p-4 border rounded-md bg-surface-subtle flex flex-col gap-y-4">
          <Input
            id="slack-channels"
            label="Slack Channels"
            hint="Comma-separated list of channels"
            bind:value={slackChannelsInput}
            onInput={handleSlackChannelsInput}
            placeholder="e.g. #alerts, #data-team"
          />

          <Input
            id="slack-users"
            label="Slack Users"
            hint="Comma-separated list of user IDs"
            bind:value={slackUsersInput}
            onInput={handleSlackUsersInput}
            placeholder="e.g. U123ABC, U456DEF"
          />

          <Input
            id="email-recipients"
            label="Email Recipients"
            hint="Comma-separated list of email addresses"
            bind:value={emailRecipientsInput}
            onInput={handleEmailRecipientsInput}
            placeholder="e.g. team@example.com, alerts@example.com"
          />
        </div>
      </div>

      <!-- Error Banner -->
      {#if hasErrors}
        <div class="error-banner">
          <AlertCircleIcon size="16px" class="flex-shrink-0 mt-0.5" />
          <div class="flex flex-col gap-y-1">
            {#each errorMessages as message}
              <span>{message}</span>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .type-button {
    @apply flex flex-col items-start p-3 rounded-[2px] border text-left transition-colors;
    @apply bg-surface-subtle;
  }

  .type-button:hover {
    @apply bg-surface-hover;
  }

  .type-button.selected {
    @apply bg-primary-50 border-primary-500;
  }

  .type-label {
    @apply text-sm font-medium;
  }

  .type-button.selected .type-label {
    @apply text-primary-700;
  }

  .type-description {
    @apply text-xs text-fg-muted mt-0.5;
  }

  .error-banner {
    @apply flex items-start gap-x-2 p-3 rounded-[2px] border;
    @apply bg-red-50 border-red-200 text-red-700 text-sm;
  }
</style>
