import { describe, expect, it } from "vitest";
import {
  getAlertOptionsFromFormValues,
  getFormValuesFromAlertSpec,
  isResourceErrorAlert,
  toggleResource,
  type ResourceErrorAlertFormValues,
} from "./utils";

const values: ResourceErrorAlertFormValues = {
  name: " Errors in orders ",
  resources: ["rill.runtime.v1.Model/orders"],
  refreshCron: "",
  refreshTimeZone: "",
  emailRecipients: ["admin@example.com"],
  slackChannels: ["#alerts", ""],
  slackUsers: [],
  snooze: "3600",
};

describe("resource error alert utils", () => {
  it("maps form values to alert options", () => {
    expect(getAlertOptionsFromFormValues(values)).toEqual({
      displayName: "Errors in orders",
      refreshCron: "",
      refreshTimeZone: "",
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
      refreshSchedule: { cron: "0 * * * *", timeZone: "Asia/Kolkata" },
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
    const formValues = getFormValuesFromAlertSpec(spec);
    expect(formValues).toEqual({
      ...values,
      name: "Errors in orders",
      refreshCron: "0 * * * *",
      refreshTimeZone: "Asia/Kolkata",
      slackChannels: ["#alerts"],
    });
    // The saved schedule is sent back on edits
    expect(getAlertOptionsFromFormValues(formValues)).toMatchObject({
      refreshCron: "0 * * * *",
      refreshTimeZone: "Asia/Kolkata",
    });
  });

  it("toggles resources with a kind wildcard exclusive of its resources", () => {
    const model = "rill.runtime.v1.Model";
    const source = "rill.runtime.v1.Source/raw";
    const list = [`${model}/orders`, `${model}/users`, source];

    // Selecting the wildcard replaces the kind's resources
    const withAll = toggleResource(list, `${model}/*`);
    expect(withAll).toEqual([source, `${model}/*`]);

    // Selecting a resource replaces the kind's wildcard
    expect(toggleResource(withAll, `${model}/orders`)).toEqual([
      source,
      `${model}/orders`,
    ]);

    // Deselecting removes only the key
    expect(toggleResource(withAll, `${model}/*`)).toEqual([source]);
    expect(toggleResource(list, source)).toEqual(list.slice(0, 2));
  });

  it("does not treat dashboard alerts as resource error alerts", () => {
    expect(isResourceErrorAlert({ annotations: {} })).toBe(false);
    expect(isResourceErrorAlert(undefined)).toBe(false);
  });
});
