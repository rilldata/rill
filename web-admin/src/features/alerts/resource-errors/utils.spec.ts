import { describe, expect, it } from "vitest";
import {
  getAlertOptionsFromFormValues,
  getFormValuesFromAlertSpec,
  isResourceErrorAlert,
  type ResourceErrorAlertFormValues,
} from "./utils";

const values: ResourceErrorAlertFormValues = {
  name: " Errors in orders ",
  resources: ["rill.runtime.v1.Model/orders"],
  refreshCron: "*/5 * * * *",
  emailRecipients: ["admin@example.com"],
  slackChannels: ["#alerts", ""],
  slackUsers: [],
  snooze: "3600",
};

describe("resource error alert utils", () => {
  it("maps form values to alert options", () => {
    expect(getAlertOptionsFromFormValues(values)).toEqual({
      displayName: "Errors in orders",
      refreshCron: "*/5 * * * *",
      resolver: "resource_status",
      resolverProperties: {
        where_error: true,
        where_resource: ["rill.runtime.v1.Model/orders"],
      },
      emailRecipients: ["admin@example.com"],
      slackChannels: ["#alerts"],
      slackUsers: [],
      renotify: true,
      renotifyAfterSeconds: 3600,
      webOpenPath: "/-/status/resources?status=error",
    });
  });

  it("omits where_resource and renotify for project-wide alerts without snooze", () => {
    const options = getAlertOptionsFromFormValues({
      ...values,
      resources: [],
      snooze: "0",
    });
    expect(options.resolverProperties).toEqual({ where_error: true });
    expect(options.renotify).toBe(false);
    expect(options.renotifyAfterSeconds).toBe(0);
  });

  it("round-trips through an alert spec", () => {
    const spec = {
      displayName: "Errors in orders",
      refreshSchedule: { cron: "*/5 * * * *" },
      resolver: "resource_status",
      resolverProperties: {
        where_error: true,
        where_resource: ["rill.runtime.v1.Model/orders"],
      },
      renotifyAfterSeconds: 3600,
      notifiers: [
        {
          connector: "email",
          properties: { recipients: ["admin@example.com"] },
        },
        { connector: "slack", properties: { channels: ["#alerts"] } },
      ],
      annotations: { admin_system: "true" },
    };
    expect(isResourceErrorAlert(spec)).toBe(true);
    expect(getFormValuesFromAlertSpec(spec)).toEqual({
      ...values,
      name: "Errors in orders",
      slackChannels: ["#alerts"],
    });
  });

  it("does not treat dashboard alerts as resource error alerts", () => {
    expect(isResourceErrorAlert({ annotations: {} })).toBe(false);
    expect(isResourceErrorAlert(undefined)).toBe(false);
  });
});
