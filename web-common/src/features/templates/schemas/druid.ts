import type { MultiStepFormSchema } from "./types";

export const druidSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "connection_string"],
      default: "parameters",
      description: "Choose how to connect to Druid",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-grouped-fields": {
        parameters: ["host", "port", "username", "password", "ssl"],
        connection_string: ["dsn"],
      },
      "x-step": "connector",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname or IP address of the Druid server",
      "x-placeholder": "localhost",
      "x-step": "connector",
      "x-visible-if": { auth_method: "parameters" },
    },
    port: {
      type: "number",
      title: "Port",
      description: "Port number of the Druid server",
      "x-placeholder": "8888",
      "x-step": "connector",
      "x-visible-if": { auth_method: "parameters" },
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to the Druid server (optional)",
      "x-placeholder": "default",
      "x-step": "connector",
      "x-visible-if": { auth_method: "parameters" },
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to the Druid server (optional)",
      "x-placeholder": "Enter password",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "parameters" },
    },
    ssl: {
      type: "boolean",
      title: "Use SSL",
      description: "Use SSL to connect to the Druid server",
      default: true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "parameters" },
    },
    dsn: {
      type: "string",
      title: "Connection String",
      description: "Druid connection string (DSN)",
      "x-placeholder": "https://example.com/druid/v2/sql/avatica-protobuf?authentication=BASIC&avaticaUser=username&avaticaPassword=password",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "connection_string" },
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "parameters" } } },
      then: { required: ["host", "ssl"] },
    },
    {
      if: { properties: { auth_method: { const: "connection_string" } } },
      then: { required: ["dsn"] },
    },
  ],
};
