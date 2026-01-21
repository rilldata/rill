import type { MultiStepFormSchema } from "./types";

export const salesforceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    soql: {
      type: "string",
      title: "SOQL",
      description: "SOQL query to extract data",
      "x-monospace": true,
      "x-placeholder": "SELECT Id, Name FROM Opportunity",
      "x-step": "source",
    },
    sobject: {
      type: "string",
      title: "SObject",
      description: "Salesforce object to query",
      "x-placeholder": "Opportunity",
      "x-step": "source",
    },
    queryAll: {
      type: "boolean",
      title: "Query all",
      description: "Include deleted and archived records",
      default: false,
      "x-step": "source",
    },
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
      "x-placeholder": "your_password_or_password+token",
      "x-secret": true,
    },
    key: {
      type: "string",
      title: "JWT private key",
      description: "PEM-formatted private key for JWT auth",
      "x-display": "textarea",
      "x-secret": true,
    },
    client_id: {
      type: "string",
      title: "Connected App Client ID",
      description: "Client ID (consumer key) for JWT auth",
      "x-placeholder": "Connected App client ID",
    },
    endpoint: {
      type: "string",
      title: "Login endpoint",
      description:
        "Salesforce login URL (e.g., login.salesforce.com or test.salesforce.com)",
      "x-placeholder": "login.salesforce.com",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source",
      "x-placeholder": "my_new_source",
      "x-step": "source",
    },
  },
  required: ["soql", "sobject", "name"],
  allOf: [
    {
      if: {
        properties: {
          key: { const: "" },
        },
      },
      then: {
        required: ["username", "password", "endpoint"],
      },
    },
    {
      if: {
        properties: {
          key: { const: undefined },
        },
      },
      then: {
        required: ["username", "password", "endpoint"],
      },
    },
  ],
};
