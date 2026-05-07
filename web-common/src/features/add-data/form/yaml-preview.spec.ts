import { describe, expect, it } from "vitest";
import { makeTestEnvEditSession } from "@rilldata/web-common/features/env-management/test/test-env-store.ts";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { getConnectorYamlPreview } from "@rilldata/web-common/features/add-data/form/yaml-preview.ts";
import { clickhouseSchema } from "@rilldata/web-common/features/templates/schemas/clickhouse.ts";
import { ducklakeSchema } from "@rilldata/web-common/features/templates/schemas/ducklake.ts";

describe("getConnectorYamlPreview", () => {
  describe("clickhouse", () => {
    const connector: V1ConnectorDriver = { name: "clickhouse" };
    const schema = clickhouseSchema;
    const formValuesWithoutPassword = {
      host: "ch.example.com",
      username: "user",
    };
    const formValuesWithPassword = {
      ...formValuesWithoutPassword,
      password: "pass",
    };

    it("should retain same value across edit commits", async () => {
      const { envEditSession, testEnvs, envStore } =
        await makeTestEnvEditSession();
      const yamlInitial = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithPassword,
        envEditSession,
      });
      expect(yamlInitial).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass",
      });

      // New changes arrived but value didnt change
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass";
      await envStore.pull();

      const yamlAfterPull = getConnectorYamlPreview({
        connector,
        schema,
        formValues: {
          ...formValuesWithPassword,
          password: "pass_1",
        },
        envEditSession,
      });
      expect(yamlAfterPull).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass_1",
      });

      // New changes arrived with new values.
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass_source";
      await envStore.pull();

      const yamlAfterSourceUpdate = getConnectorYamlPreview({
        connector,
        schema,
        formValues: {
          ...formValuesWithPassword,
          password: "pass_2",
        },
        envEditSession,
      });
      expect(yamlAfterSourceUpdate).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD_1 }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass_source",
        CLICKHOUSE_PASSWORD_1: "pass_2",
      });
    });

    it("should delete unused vars if not updated from outside", async () => {
      const { envEditSession, testEnvs, envStore } =
        await makeTestEnvEditSession();
      const yamlInitial = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithPassword,
        envEditSession,
      });
      expect(yamlInitial).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass",
      });

      // New changes arrived but value didnt change
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass";
      await envStore.pull();

      const yamlWithoutPassword = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithoutPassword,
        envEditSession,
      });
      expect(yamlWithoutPassword).not.toContain("password:");
      await envEditSession.commit();
      expect(testEnvs).toEqual({});
    });

    it("should delete vars on rollback if not updated from outside", async () => {
      const { envEditSession, testEnvs, envStore } =
        await makeTestEnvEditSession();
      const yamlInitial = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithPassword,
        envEditSession,
      });
      expect(yamlInitial).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass",
      });

      // New changes arrived but value didnt change
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass";
      await envStore.pull();

      await envEditSession.rollback();
      expect(testEnvs).toEqual({});
    });

    it("should retain unused vars if updated from outside", async () => {
      const { envEditSession, testEnvs, envStore } =
        await makeTestEnvEditSession();
      const yamlInitial = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithPassword,
        envEditSession,
      });
      expect(yamlInitial).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass",
      });

      // New changes arrived with new values.
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass_source";
      await envStore.pull();

      const yamlWithoutPassword = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithoutPassword,
        envEditSession,
      });
      expect(yamlWithoutPassword).not.toContain("password:");
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass_source",
      });
    });

    it("should retain vars on rollback if updated from outside", async () => {
      const { envEditSession, testEnvs, envStore } =
        await makeTestEnvEditSession();
      const yamlInitial = getConnectorYamlPreview({
        connector,
        schema,
        formValues: formValuesWithPassword,
        envEditSession,
      });
      expect(yamlInitial).toContain(
        `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
      );
      await envEditSession.commit();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass",
      });

      // New changes arrived with new values.
      testEnvs["CLICKHOUSE_PASSWORD"] = "pass_source";
      await envStore.pull();

      await envEditSession.rollback();
      expect(testEnvs).toEqual({
        CLICKHOUSE_PASSWORD: "pass_source",
      });
    });
  });

  describe("ducklake", () => {
    const connector: V1ConnectorDriver = { name: "ducklake" };
    const schema = ducklakeSchema;

    describe("direct attach field", () => {
      it("should retain same value across edit commits", async () => {
        const { envEditSession, testEnvs, envStore } =
          await makeTestEnvEditSession();
        const yamlInitial = getConnectorYamlPreview({
          connector,
          schema,
          formValues: {
            connection_mode: "sql",
            attach:
              "'ducklake:postgres:dbname=mydb host=localhost user=postgres password=pass'",
          },
          envEditSession,
        });
        expect(yamlInitial).toContain(
          `attach: "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_POSTGRES:
            "dbname=mydb host=localhost user=postgres password=pass",
        });

        // New changes arrived but value didnt change
        testEnvs["DUCKLAKE_POSTGRES"] =
          "dbname=mydb host=localhost user=postgres password=pass";
        await envStore.pull();

        const yamlAfterPull = getConnectorYamlPreview({
          connector,
          schema,
          formValues: {
            connection_mode: "sql",
            attach:
              "'ducklake:postgres:dbname=mydb host=localhost user=postgres password=pass_1'",
          },
          envEditSession,
        });
        expect(yamlAfterPull).toContain(
          `attach: "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_POSTGRES:
            "dbname=mydb host=localhost user=postgres password=pass_1",
        });

        // New changes arrived with new values.
        testEnvs["DUCKLAKE_POSTGRES"] =
          "dbname=mydb host=localhost user=postgres password=pass_source";
        await envStore.pull();

        const yamlAfterSourceUpdate = getConnectorYamlPreview({
          connector,
          schema,
          formValues: {
            connection_mode: "sql",
            attach:
              "'ducklake:postgres:dbname=mydb host=localhost user=postgres password=pass_2'",
          },
          envEditSession,
        });
        expect(yamlAfterSourceUpdate).toContain(
          `attach: "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES_1 }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_POSTGRES:
            "dbname=mydb host=localhost user=postgres password=pass_source",
          DUCKLAKE_POSTGRES_1:
            "dbname=mydb host=localhost user=postgres password=pass_2",
        });
      });
    });

    describe("build attach from params", () => {
      const formValuesWithoutPassword = {
        connection_mode: "parameters",
        catalog_type: "postgres",
        catalog_postgres_dbname: "mydb",
        catalog_postgres_host: "localhost",
        catalog_postgres_user: "postgres",
      };
      const formValuesWithPassword = {
        ...formValuesWithoutPassword,
        catalog_postgres_password: "pass",
      };

      it("should retain same value across edit commits for separate fields", async () => {
        const { envEditSession, testEnvs, envStore } =
          await makeTestEnvEditSession();
        const yamlInitial = getConnectorYamlPreview({
          connector,
          schema,
          formValues: formValuesWithPassword,
          envEditSession,
        });
        expect(yamlInitial).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass",
        });

        // New changes arrived but value didnt change
        testEnvs["DUCKLAKE_CATALOG_POSTGRES_PASSWORD"] = "pass";
        await envStore.pull();

        const yamlAfterPull = getConnectorYamlPreview({
          connector,
          schema,
          formValues: {
            ...formValuesWithoutPassword,
            catalog_postgres_password: "pass_1",
          },
          envEditSession,
        });
        expect(yamlAfterPull).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass_1",
        });

        // New changes arrived with new values.
        testEnvs["DUCKLAKE_CATALOG_POSTGRES_PASSWORD"] = "pass_source";
        await envStore.pull();

        const yamlAfterSourceUpdate = getConnectorYamlPreview({
          connector,
          schema,
          formValues: {
            ...formValuesWithoutPassword,
            catalog_postgres_password: "pass_2",
          },
          envEditSession,
        });
        expect(yamlAfterSourceUpdate).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD_1 }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass_source",
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD_1: "pass_2",
        });
      });

      it("should delete unused vars if not updated from outside", async () => {
        const { envEditSession, testEnvs, envStore } =
          await makeTestEnvEditSession();
        const yamlInitial = getConnectorYamlPreview({
          connector,
          schema,
          formValues: formValuesWithPassword,
          envEditSession,
        });
        expect(yamlInitial).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass",
        });

        // New changes arrived but value didnt change
        testEnvs["DUCKLAKE_CATALOG_POSTGRES_PASSWORD"] = "pass";
        await envStore.pull();

        const yamlWithoutPassword = getConnectorYamlPreview({
          connector,
          schema,
          formValues: formValuesWithoutPassword,
          envEditSession,
        });
        expect(yamlWithoutPassword).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({});
      });

      it("should retain unused vars if updated from outside", async () => {
        const { envEditSession, testEnvs, envStore } =
          await makeTestEnvEditSession();
        const yamlInitial = getConnectorYamlPreview({
          connector,
          schema,
          formValues: formValuesWithPassword,
          envEditSession,
        });
        expect(yamlInitial).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass",
        });

        // New changes arrived with new values.
        testEnvs["DUCKLAKE_CATALOG_POSTGRES_PASSWORD"] = "pass_source";
        await envStore.pull();

        const yamlWithoutPassword = getConnectorYamlPreview({
          connector,
          schema,
          formValues: formValuesWithoutPassword,
          envEditSession,
        });
        expect(yamlWithoutPassword).toContain(
          `attach: "'ducklake:postgres:dbname=mydb host=localhost user=postgres'"`,
        );
        await envEditSession.commit();
        expect(testEnvs).toEqual({
          DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "pass_source",
        });
      });
    });
  });
});
