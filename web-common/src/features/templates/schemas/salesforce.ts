import type { MultiStepFormSchema } from "./types";

export const salesforceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Salesforce",
  "x-category": "warehouse",
  properties: {
    soql: {
      type: "string",
      title: "SOQL",
      description: "SOQL query to extract data",
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
      "x-env-var-name": "SALESFORCE_PASSWORD",
    },
    key: {
      type: "string",
      title: "JWT private key",
      description: "PEM-formatted private key for JWT auth",
      "x-display": "textarea",
      "x-placeholder": "your_private_key",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_KEY",
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
      // Username/password auth: when key is NOT provided, require username/password/endpoint
      if: {
        not: {
          required: ["key"],
          properties: {
            key: { minLength: 1 },
          },
        },
      },
      then: {
        required: ["username", "password", "endpoint"],
      },
    },
    {
      // JWT auth: when key is provided, require client_id and username
      if: {
        required: ["key"],
        properties: {
          key: { minLength: 1 },
        },
      },
      then: {
        required: ["client_id", "username"],
      },
    },
  ],
};
