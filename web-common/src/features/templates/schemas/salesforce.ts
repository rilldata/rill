import type { MultiStepFormSchema } from "./types";

export const salesforceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    username: {
      type: "string",
      title: "Username",
      description: "Salesforce username",
      "x-placeholder": "Enter your Salesforce username",
      "x-step": "source",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Salesforce password",
      "x-placeholder": "Enter your Salesforce password",
      "x-secret": true,
      "x-step": "source",
    },
    key: {
      type: "string",
      title: "Security Token",
      description: "Salesforce security token",
      "x-placeholder": "Enter your security token",
      "x-secret": true,
      "x-step": "source",
    },
    endpoint: {
      type: "string",
      title: "Endpoint",
      description: "Salesforce endpoint URL (optional, defaults to production)",
      "x-placeholder": "https://login.salesforce.com",
      "x-step": "source",
      "x-advanced": true,
    },
    client_id: {
      type: "string",
      title: "Client ID",
      description: "Connected App client ID (optional)",
      "x-placeholder": "Enter client ID",
      "x-step": "source",
      "x-advanced": true,
    },
    soql: {
      type: "string",
      title: "SOQL Query",
      description: "SOQL Query to extract data from Salesforce",
      "x-placeholder": "SELECT Id, CreatedDate, Name FROM Opportunity",
      "x-hint":
        "Write a SOQL query to retrieve data from your Salesforce object.",
      "x-step": "source",
    },
    sobject: {
      type: "string",
      title: "SObject",
      description: "SObject to query in Salesforce",
      "x-placeholder": "Opportunity",
      "x-hint":
        "Enter the name of the Salesforce object you want to query (e.g., Opportunity, Lead, Account).",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_salesforce_source",
      "x-step": "source",
    },
  },
  required: ["soql", "sobject", "name"],
};
