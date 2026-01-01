import type { MultiStepFormSchema } from "./types";

export const pinotSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "Full Pinot connection string, e.g. http(s)://user:password@broker:8000?controller=host:9000",
      "x-placeholder":
        "https://username:password@localhost:8000?controller=localhost:9000",
      "x-secret": true,
    },
    broker_host: {
      type: "string",
      title: "Broker host",
      description: "Pinot broker host",
      "x-placeholder": "localhost",
    },
    broker_port: {
      type: "string",
      title: "Broker port",
      description: "Pinot broker port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      "x-placeholder": "8000",
    },
    controller_host: {
      type: "string",
      title: "Controller host",
      description: "Pinot controller host",
      "x-placeholder": "localhost",
    },
    controller_port: {
      type: "string",
      title: "Controller port",
      description: "Pinot controller port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      "x-placeholder": "9000",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Pinot username",
      "x-placeholder": "default",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Pinot password",
      "x-placeholder": "password",
      "x-secret": true,
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL",
      default: true,
    },
  },
  required: [],
  oneOf: [
    {
      title: "Use connection string",
      required: ["dsn"],
    },
    {
      title: "Use individual parameters",
      required: ["broker_host", "controller_host", "ssl"],
    },
  ],
};
