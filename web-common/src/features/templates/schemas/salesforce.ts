import type { MultiStepFormSchema } from "./types";

export const salesforceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Salesforce",
  "x-category": "fileStore",
  "x-form-height": "medium",
  properties: {
    username: {
      type: "string",
      title: "Username",
      description: "Salesforce username (usually an email)",
      "x-placeholder": "user@example.com",
    },
    password: {
      type: "string",
      title: "Password",
      description:
        "Salesforce password, optionally followed by security token if required",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_PASSWORD",
    },
    endpoint: {
      type: "string",
      title: "Login endpoint",
      description:
        "Salesforce login URL (e.g., login.salesforce.com or test.salesforce.com)",
      "x-placeholder": "login.salesforce.com",
    },
    key: {
      type: "string",
      title: "JWT private key",
      description: "PEM-formatted private key for JWT auth",
      "x-display": "textarea",
      "x-placeholder": "your_private_key",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_KEY",
      "x-advanced": true,
    },
    client_id: {
      type: "string",
      title: "Connected App Client ID",
      description: "Client ID (consumer key) for JWT auth",
      "x-placeholder": "Connected App client ID",
      "x-advanced": true,
    },
    soql: {
      type: "string",
      title: "SOQL",
      description: "SOQL query to extract data",
      "x-placeholder": "SELECT Id, Name FROM Opportunity",
    },
    sobject: {
      type: "string",
      title: "SObject",
      description: "Salesforce object to query",
      "x-placeholder": "Opportunity",
    },
    queryAll: {
      type: "boolean",
      title: "Query all",
      description: "Include deleted and archived records",
      default: false,
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_new_source",
    },
  },
  required: ["soql", "sobject", "name", "username", "password"],
};
