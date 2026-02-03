<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { parseDocument, Document } from "yaml";
  import AlertVisualEditor from "./visual-editor/AlertVisualEditor.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import {
    MeasureFilterOperation,
    MeasureFilterType,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { V1Operation } from "@rilldata/web-common/runtime-client";

  export let fileArtifact: FileArtifact;
  export let alertName: string;
  export let errors: LineStatus[] = [];

  $: ({ editorContent, updateEditorContent } = fileArtifact);

  // Store the current YAML document for in-place modifications
  let currentYamlDoc: Document | null = null;
  let lastParsedContent = "";

  // Parse the YAML content to extract alert properties
  $: parsedAlert = parseAlertYaml($editorContent ?? "");

  // Update the stored YAML document when content changes externally
  $: if ($editorContent && $editorContent !== lastParsedContent) {
    try {
      currentYamlDoc = parseDocument($editorContent);
      lastParsedContent = $editorContent;
    } catch {
      // Keep the old document if parsing fails
    }
  }

  function parseAlertYaml(content: string) {
    try {
      const doc = parseDocument(content);
      const data = doc.toJSON() ?? {};
      return {
        name: data.display_name ?? data.name ?? alertName ?? "",
        measure: data.measure ?? "",
        splitByDimension: data.split_by_dimension ?? "",
        criteria: parseCriteria(data.data?.criteria ?? []),
        criteriaOperation: data.criteria_operation === "or"
          ? V1Operation.OPERATION_OR
          : V1Operation.OPERATION_AND,
        snooze: String(data.renotify_after_seconds ?? ""),
        enableSlackNotification: !!(data.notify?.slack?.channels?.length || data.notify?.slack?.users?.length),
        slackChannels: data.notify?.slack?.channels ?? [],
        slackUsers: data.notify?.slack?.users ?? [],
        enableEmailNotification: !!data.notify?.email?.recipients?.length,
        emailRecipients: data.notify?.email?.recipients ?? [],
        refreshWhenDataRefreshes: data.refresh?.ref_update ?? !data.refresh?.cron,
        frequency: "Daily",
        dayOfWeek: "Monday",
        timeOfDay: "09:00",
        dayOfMonth: "1",
        timeZone: data.time_zone ?? "UTC",
        // Data source fields (under data: key)
        metricsSql: data.data?.metrics_sql ?? "",
        sql: data.data?.sql ?? "",
        sqlConnector: data.data?.connector ?? "",
        resourceError: "",
        whenErrors: data.data?.resource_status?.where_error ?? false,
        cron: data.refresh?.cron ?? "",
      };
    } catch {
      return {
        name: alertName ?? "",
        measure: "",
        splitByDimension: "",
        criteria: [] as MeasureFilterEntry[],
        criteriaOperation: V1Operation.OPERATION_AND,
        snooze: "",
        enableSlackNotification: false,
        slackChannels: [] as string[],
        slackUsers: [] as string[],
        enableEmailNotification: false,
        emailRecipients: [] as string[],
        refreshWhenDataRefreshes: true,
        frequency: "Daily",
        dayOfWeek: "Monday",
        timeOfDay: "09:00",
        dayOfMonth: "1",
        timeZone: "UTC",
        metricsSql: "",
        sql: "",
        sqlConnector: "",
        resourceError: "",
        whenErrors: false,
        cron: "",
      };
    }
  }

  function parseCriteria(criteria: any[]): MeasureFilterEntry[] {
    if (!Array.isArray(criteria)) return [];
    return criteria.map((c) => ({
      measure: c.field ?? c.measure ?? "",
      type: c.type ?? MeasureFilterType.Value,
      operation: c.operation ?? MeasureFilterOperation.GreaterThan,
      value1: String(c.value ?? c.value1 ?? ""),
      value2: String(c.value2 ?? ""),
    }));
  }

  // Local state for editing - initialized from defaults, updated reactively
  let name = "";
  let measure = "";
  let splitByDimension = "";
  let criteria: MeasureFilterEntry[] = [];
  let criteriaOperation: V1Operation = V1Operation.OPERATION_AND;
  let snooze = "";
  let enableSlackNotification = false;
  let slackChannels: string[] = [];
  let slackUsers: string[] = [];
  let enableEmailNotification = false;
  let emailRecipients: string[] = [];
  let refreshWhenDataRefreshes = true;
  let frequency = "Daily";
  let dayOfWeek = "Monday";
  let timeOfDay = "09:00";
  let dayOfMonth = "1";
  let timeZone = "UTC";
  // Data source fields
  let metricsSql = "";
  let sql = "";
  let sqlConnector = "";
  let resourceError = "";
  let whenErrors = false;
  let cron = "";
  let initialized = false;

  // Update local state when parsed alert changes (due to external YAML edits)
  $: if (parsedAlert && !initialized) {
    name = parsedAlert.name;
    measure = parsedAlert.measure;
    splitByDimension = parsedAlert.splitByDimension;
    criteria = parsedAlert.criteria;
    criteriaOperation = parsedAlert.criteriaOperation;
    snooze = parsedAlert.snooze;
    enableSlackNotification = parsedAlert.enableSlackNotification;
    slackChannels = parsedAlert.slackChannels;
    slackUsers = parsedAlert.slackUsers;
    enableEmailNotification = parsedAlert.enableEmailNotification;
    emailRecipients = parsedAlert.emailRecipients;
    refreshWhenDataRefreshes = parsedAlert.refreshWhenDataRefreshes;
    frequency = parsedAlert.frequency;
    dayOfWeek = parsedAlert.dayOfWeek;
    timeOfDay = parsedAlert.timeOfDay;
    dayOfMonth = parsedAlert.dayOfMonth;
    timeZone = parsedAlert.timeZone;
    metricsSql = parsedAlert.metricsSql;
    sql = parsedAlert.sql;
    sqlConnector = parsedAlert.sqlConnector;
    resourceError = parsedAlert.resourceError;
    whenErrors = parsedAlert.whenErrors;
    cron = parsedAlert.cron;
    initialized = true;
  }

  // Auto-save to YAML when form values change
  // Create a derived value that depends on all form fields
  $: formState = {
    name,
    measure,
    splitByDimension,
    criteria,
    criteriaOperation,
    snooze,
    enableSlackNotification,
    slackChannels,
    slackUsers,
    enableEmailNotification,
    emailRecipients,
    refreshWhenDataRefreshes,
    frequency,
    dayOfWeek,
    timeOfDay,
    dayOfMonth,
    timeZone,
    metricsSql,
    sql,
    sqlConnector,
    resourceError,
    whenErrors,
    cron,
  };

  // React to formState changes and auto-save
  $: if (initialized && formState && currentYamlDoc) {
    const yamlContent = updateAlertYaml();
    if (yamlContent !== lastParsedContent) {
      void updateEditorContent(yamlContent, false, true);
    }
  }

  // Helper to set nested values in YAML document
  function setNestedValue(doc: Document, path: string[], value: any) {
    if (!doc.contents) return;
    let current: any = doc.contents;
    for (let i = 0; i < path.length - 1; i++) {
      const next = current.get(path[i], true);
      if (!next) return; // Don't create new paths
      current = next;
    }
    if (current && current.has && current.has(path[path.length - 1])) {
      current.set(path[path.length - 1], value);
    }
  }

  function updateAlertYaml(): string {
    if (!currentYamlDoc) return lastParsedContent;

    // Clone the document to avoid mutating the original
    const doc = parseDocument(currentYamlDoc.toString());
    const contents = doc.contents as any;

    // Update only the fields that exist in the YAML
    setNestedValue(doc, ["display_name"], name);
    setNestedValue(doc, ["refresh", "cron"], cron);
    setNestedValue(doc, ["refresh", "ref_update"], refreshWhenDataRefreshes);

    // Handle data source - need to switch between types
    if (contents && contents.has && contents.has("data")) {
      const dataNode = contents.get("data", true);
      if (dataNode && dataNode.has && dataNode.set && dataNode.delete) {
        // Clear all data source types first
        try {
          if (dataNode.has("metrics_sql")) dataNode.delete("metrics_sql");
          if (dataNode.has("sql")) dataNode.delete("sql");
          if (dataNode.has("connector")) dataNode.delete("connector");
          if (dataNode.has("resource_status")) dataNode.delete("resource_status");

          // Set the active data source
          if (metricsSql) {
            dataNode.set("metrics_sql", metricsSql);
          } else if (sql) {
            if (sqlConnector) dataNode.set("connector", sqlConnector);
            dataNode.set("sql", sql);
          } else if (selectedDataSource === "resource_error") {
            dataNode.set("resource_status", { where_error: whenErrors });
          }
        } catch (e) {
          console.error("Error updating data node:", e);
        }
      }
    }

    setNestedValue(doc, ["notify", "slack", "channels"], slackChannels);
    setNestedValue(doc, ["notify", "slack", "users"], slackUsers);
    setNestedValue(doc, ["notify", "email", "recipients"], emailRecipients);

    return doc.toString();
  }

  // Mock metrics view spec - in a real implementation this would come from context
  const metricsViewSpec = {
    measures: [],
    dimensions: [],
  };
</script>

<div class="h-full w-full bg-surface-subtle overflow-auto">
  <AlertVisualEditor
    bind:name
    bind:measure
    bind:splitByDimension
    bind:criteria
    bind:criteriaOperation
    bind:snooze
    bind:enableSlackNotification
    bind:slackChannels
    bind:slackUsers
    bind:enableEmailNotification
    bind:emailRecipients
    bind:refreshWhenDataRefreshes
    bind:frequency
    bind:dayOfWeek
    bind:timeOfDay
    bind:dayOfMonth
    bind:timeZone
    bind:metricsSql
    bind:sql
    bind:sqlConnector
    bind:resourceError
    bind:whenErrors
    bind:cron
    {metricsViewSpec}
    {errors}
    hasTimeComparison={false}
  />
</div>
