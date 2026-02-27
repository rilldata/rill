import ApachePinot from "../../../components/icons/connectors/ApachePinot.svelte";
import ApachePinotIcon from "../../../components/icons/connectors/ApachePinotIcon.svelte";
import type { MultiStepFormSchema } from "./types";

export const pinotSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Apache Pinot",
  "x-category": "olap",
  "x-icon": ApachePinot,
  "x-small-icon": ApachePinotIcon,
  properties: {
    connection_mode: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-ui-only": true,
      "x-tab-group": {
        parameters: [
          "broker_host",
          "broker_port",
          "controller_host",
          "controller_port",
          "username",
          "password",
          "ssl",
        ],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "Full Pinot connection string, e.g. http(s)://user:password@broker:8000?controller=host:9000",
      "x-placeholder":
        "https://username:password@localhost:8000?controller=localhost:9000",
      "x-secret": true,
      "x-env-var-name": "PINOT_DSN",
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
      "x-env-var-name": "PINOT_PASSWORD",
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
