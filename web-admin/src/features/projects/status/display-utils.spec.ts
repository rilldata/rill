import { describe, it, expect } from "vitest";
import { V1DeploymentStatus } from "@rilldata/web-admin/client";
import {
  formatEnvironmentName,
  getStatusDotClass,
  getStatusLabel,
  formatConnectorName,
  getResourceKindTagColor,
} from "./display-utils";

describe("display-utils", () => {
  describe("formatEnvironmentName", () => {
    it("returns 'Production' for undefined", () => {
      expect(formatEnvironmentName(undefined)).toBe("Production");
    });

    it("returns 'Production' for empty string", () => {
      expect(formatEnvironmentName("")).toBe("Production");
    });

    it("handles 'prod' case-insensitively", () => {
      expect(formatEnvironmentName("prod")).toBe("Production");
      expect(formatEnvironmentName("PROD")).toBe("Production");
      expect(formatEnvironmentName("Prod")).toBe("Production");
    });

    it("handles 'production' case-insensitively", () => {
      expect(formatEnvironmentName("production")).toBe("Production");
      expect(formatEnvironmentName("PRODUCTION")).toBe("Production");
    });

    it("handles 'dev' case-insensitively", () => {
      expect(formatEnvironmentName("dev")).toBe("Development");
      expect(formatEnvironmentName("DEV")).toBe("Development");
    });

    it("handles 'development' case-insensitively", () => {
      expect(formatEnvironmentName("development")).toBe("Development");
      expect(formatEnvironmentName("DEVELOPMENT")).toBe("Development");
    });

    it("handles 'stage' case-insensitively", () => {
      expect(formatEnvironmentName("stage")).toBe("Staging");
      expect(formatEnvironmentName("STAGE")).toBe("Staging");
    });

    it("handles 'staging' case-insensitively", () => {
      expect(formatEnvironmentName("staging")).toBe("Staging");
      expect(formatEnvironmentName("STAGING")).toBe("Staging");
    });

    it("capitalizes first letter for other environments", () => {
      expect(formatEnvironmentName("test")).toBe("Test");
      expect(formatEnvironmentName("qa")).toBe("Qa");
      expect(formatEnvironmentName("preview")).toBe("Preview");
    });
  });

  describe("getStatusDotClass", () => {
    it("returns green for RUNNING status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING),
      ).toBe("bg-green-500");
    });

    it("returns yellow for PENDING status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING),
      ).toBe("bg-yellow-500");
    });

    it("returns yellow for UPDATING status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING),
      ).toBe("bg-yellow-500");
    });

    it("returns yellow for STOPPING status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING),
      ).toBe("bg-yellow-500");
    });

    it("returns yellow for DELETING status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING),
      ).toBe("bg-yellow-500");
    });

    it("returns red for ERRORED status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED),
      ).toBe("bg-red-500");
    });

    it("returns gray for UNSPECIFIED status", () => {
      expect(
        getStatusDotClass(V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED),
      ).toBe("bg-gray-400");
    });

    it("returns gray for unknown status", () => {
      expect(getStatusDotClass("unknown" as V1DeploymentStatus)).toBe(
        "bg-gray-400",
      );
    });
  });

  describe("getStatusLabel", () => {
    it("returns 'Ready' for RUNNING status", () => {
      expect(getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING)).toBe(
        "Ready",
      );
    });

    it("returns 'Pending' for PENDING status", () => {
      expect(getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING)).toBe(
        "Pending",
      );
    });

    it("returns 'Updating' for UPDATING status", () => {
      expect(
        getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING),
      ).toBe("Updating");
    });

    it("returns 'Stopping' for STOPPING status", () => {
      expect(
        getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING),
      ).toBe("Stopping");
    });

    it("returns 'Deleting' for DELETING status", () => {
      expect(
        getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING),
      ).toBe("Deleting");
    });

    it("returns 'Error' for ERRORED status", () => {
      expect(getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED)).toBe(
        "Error",
      );
    });

    it("returns 'Stopped' for STOPPED status", () => {
      expect(getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED)).toBe(
        "Stopped",
      );
    });

    it("returns 'Deleted' for DELETED status", () => {
      expect(getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED)).toBe(
        "Deleted",
      );
    });

    it("returns 'Not deployed' for UNSPECIFIED status", () => {
      expect(
        getStatusLabel(V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED),
      ).toBe("Not deployed");
    });

    it("returns 'Not deployed' for unknown status", () => {
      expect(getStatusLabel("unknown" as V1DeploymentStatus)).toBe(
        "Not deployed",
      );
    });
  });

  describe("formatConnectorName", () => {
    it("returns em dash for undefined", () => {
      expect(formatConnectorName(undefined)).toBe("—");
    });

    it("formats 'duckdb' correctly", () => {
      expect(formatConnectorName("duckdb")).toBe("DuckDB");
    });

    it("formats 'clickhouse' correctly", () => {
      expect(formatConnectorName("clickhouse")).toBe("ClickHouse");
    });

    it("formats 'druid' correctly", () => {
      expect(formatConnectorName("druid")).toBe("Druid");
    });

    it("formats 'pinot' correctly", () => {
      expect(formatConnectorName("pinot")).toBe("Pinot");
    });

    it("formats 'openai' correctly", () => {
      expect(formatConnectorName("openai")).toBe("OpenAI");
    });

    it("formats 'claude' correctly", () => {
      expect(formatConnectorName("claude")).toBe("Claude");
    });

    it("formats 'mysql' correctly", () => {
      expect(formatConnectorName("mysql")).toBe("MySQL");
    });

    it("formats 'bigquery' correctly", () => {
      expect(formatConnectorName("bigquery")).toBe("BigQuery");
    });

    it("capitalizes first letter for unknown connectors", () => {
      expect(formatConnectorName("postgres")).toBe("Postgres");
    });
  });

  describe("getResourceKindTagColor", () => {
    it("returns blue for MetricsView", () => {
      expect(getResourceKindTagColor("rill.runtime.v1.MetricsView")).toBe(
        "blue",
      );
    });

    it("returns green for Model", () => {
      expect(getResourceKindTagColor("rill.runtime.v1.Model")).toBe("green");
    });

    it("returns orange for Report", () => {
      expect(getResourceKindTagColor("rill.runtime.v1.Report")).toBe("orange");
    });

    it("returns purple for Source", () => {
      expect(getResourceKindTagColor("rill.runtime.v1.Source")).toBe("purple");
    });

    it("returns magenta for Theme", () => {
      expect(getResourceKindTagColor("rill.runtime.v1.Theme")).toBe("magenta");
    });

    it("returns gray for unknown kinds", () => {
      expect(getResourceKindTagColor("unknown")).toBe("gray");
      expect(getResourceKindTagColor("rill.runtime.v1.Unknown")).toBe("gray");
    });
  });
});
