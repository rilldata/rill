import type { MultiStepFormSchema } from "./types";

export const salesforceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Salesforce",
  "x-category": "warehouse",
  "x-form-height": "tall",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["username_password", "client_credentials", "jwt"],
      default: "username_password",
      description: "Choose how to authenticate to Salesforce",
      "x-display": "radio",
      "x-enum-labels": [
        "Username / Password (OAuth)",
        "Client Credentials",
        "JWT Bearer",
      ],
      "x-enum-descriptions": [
        "Sign in with a Salesforce username and password using the OAuth password flow. Requires a Connected App's Client ID and Client Secret (External Client Apps do not support this flow).",
        "Sign in as the run-as user using the OAuth client credentials flow. Requires a Connected App or External Client App's Client ID and Client Secret.",
        "Sign in with a JWT signed by a private key. Requires a Connected App or External Client App's Client ID and a PEM-formatted private key.",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        username_password: [
          "username",
          "password",
          "client_id",
          "client_secret",
        ],
        client_credentials: ["client_id", "client_secret"],
        jwt: ["username", "client_id", "key"],
      },
    },
    endpoint: {
      type: "string",
      title: "Login endpoint",
      description:
        "Salesforce login URL (e.g., login.salesforce.com or test.salesforce.com)",
      "x-placeholder": "login.salesforce.com",
      default: "login.salesforce.com",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Salesforce username (usually an email)",
      "x-placeholder": "user@example.com",
      "x-visible-if": { auth_method: ["username_password", "jwt"] },
    },
    password: {
      type: "string",
      title: "Password",
      description:
        "Salesforce password, optionally followed by security token if required",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_PASSWORD",
      "x-visible-if": { auth_method: "username_password" },
    },
    client_id: {
      type: "string",
      title: "Connected App Client ID",
      description:
        "Client ID for the Salesforce Connected App. The client credentials and JWT flows also accept an External Client App's Client ID.",
      "x-placeholder": "Connected App client ID",
    },
    client_secret: {
      type: "string",
      title: "Connected App Client Secret",
      description:
        "Client Secret for the Salesforce Connected App. The client credentials flow also accepts an External Client App's Client Secret.",
      "x-placeholder": "Connected App client secret",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_CLIENT_SECRET",
      "x-visible-if": {
        auth_method: ["username_password", "client_credentials"],
      },
    },
    key: {
      type: "string",
      title: "JWT private key",
      description:
        "PEM-formatted private key for JWT auth. The file is base64-encoded before being stored in .env so its newlines do not break parsing.",
      format: "file",
      "x-display": "file",
      "x-file-accept": ".pem,.key",
      "x-file-encoding": "base64",
      "x-secret": true,
      "x-env-var-name": "SALESFORCE_KEY",
      "x-visible-if": { auth_method: "jwt" },
    },
    soql: {
      type: "string",
      title: "SOQL",
      description: "SOQL query to extract data from Salesforce",
      "x-placeholder": "SELECT Id, Name FROM Opportunity",
      "x-step": "explorer",
    },
    sobject: {
      type: "string",
      title: "SObject",
      description:
        "Salesforce object the SOQL query reads from (e.g. Opportunity, Account, MyObject__c)",
      "x-placeholder": "Opportunity",
      "x-step": "explorer",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "explorer",
    },
  },
  required: ["soql", "sobject", "name"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "username_password" } } },
      then: {
        required: ["username", "password", "client_id", "client_secret"],
      },
    },
    {
      if: { properties: { auth_method: { const: "client_credentials" } } },
      then: { required: ["client_id", "client_secret"] },
    },
    {
      if: { properties: { auth_method: { const: "jwt" } } },
      then: { required: ["username", "client_id", "key"] },
    },
  ],
};
