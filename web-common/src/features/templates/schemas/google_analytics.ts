import type { MultiStepFormSchema } from "./types";

export const googleAnalyticsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Google Analytics 4",
  "x-category": "warehouse",
  "x-form-height": "tall",
  properties: {
    google_application_credentials: {
      type: "string",
      title: "GCP credentials",
      description: "Service account JSON with GA4 read access",
      format: "file",
      "x-display": "file",
      "x-file-accept": ".json",
      "x-file-encoding": "json",
      "x-secret": true,
      "x-env-var-name": "GOOGLE_APPLICATION_CREDENTIALS",
      "x-step": "connector",
    },
    property_id: {
      type: "string",
      title: "Property ID",
      description: "GA4 property ID to query",
      "x-placeholder": "123456789",
      "x-hint": "Found in Google Analytics under Admin → Property Settings.",
      "x-step": "connector",
    },
    report_type: {
      type: "string",
      title: "Report type",
      description: "Choose a predefined report or create a custom one",
      enum: ["traffic", "pages", "demographics", "events", "custom"],
      default: "traffic",
      "x-display": "select",
      "x-enum-labels": [
        "Traffic overview",
        "Page views",
        "Demographics",
        "Events",
        "Custom",
      ],
      "x-enum-descriptions": [
        "Sessions, users, bounce rate by source/medium/campaign",
        "Page views and session duration by page path/title",
        "Active users and sessions by country, city, language",
        "Event counts by event name",
        "Specify your own dimensions and metrics",
      ],
      "x-step": "explorer",
    },
    dimensions: {
      type: "string",
      title: "Dimensions",
      description: "Comma-separated GA4 dimension names",
      "x-placeholder": "date,country,city",
      "x-hint": "See GA4 API dimensions reference for available names.",
      "x-docs-url":
        "https://developers.google.com/analytics/devguides/reporting/data/v1/api-schema#dimensions",
      "x-visible-if": { report_type: "custom" },
      "x-step": "explorer",
    },
    metrics: {
      type: "string",
      title: "Metrics",
      description: "Comma-separated GA4 metric names",
      "x-placeholder": "activeUsers,sessions",
      "x-hint": "See GA4 API metrics reference for available names.",
      "x-docs-url":
        "https://developers.google.com/analytics/devguides/reporting/data/v1/api-schema#metrics",
      "x-visible-if": { report_type: "custom" },
      "x-step": "explorer",
    },
    start_date: {
      type: "string",
      title: "Start date",
      description: "Start date for the report",
      "x-placeholder": "2024-01-01",
      "x-hint": "YYYY-MM-DD format, or relative like '30daysAgo'.",
      "x-step": "explorer",
    },
    end_date: {
      type: "string",
      title: "End date",
      description: "End date for the report",
      "x-placeholder": "today",
      "x-hint": "Leave empty to default to today.",
      "x-step": "explorer",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_ga4_source",
      "x-step": "explorer",
    },
  },
  required: [
    "google_application_credentials",
    "property_id",
    "report_type",
    "start_date",
    "name",
  ],
  allOf: [
    {
      if: {
        properties: {
          report_type: { const: "custom" },
        },
      },
      then: {
        required: ["dimensions", "metrics"],
      },
    },
  ],
};
