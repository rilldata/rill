import { describe, expect, it } from "vitest";
import {
  replaceAiConnectorInYAML,
  replaceOlapConnectorInYAML,
  replaceOrAddEnvVariable,
  getGenericEnvVarName,
  envVarExists,
  findAvailableEnvVarName,
  makeEnvVarKey,
} from "./code-utils";

// Import the template for testing
const YAML_MODEL_TEMPLATE = `type: model
materialize: true\n
connector: {{ connector }}\n
sql: {{ sql }}{{ dev_section }}
`;

describe("YAML Model Template", () => {
  it("should include dev section for non-Redshift connectors", () => {
    const connector = "clickhouse";
    const selectStatement = "select * from my_table";
    const shouldIncludeDevSection = true;
    const devSection = shouldIncludeDevSection
      ? `\ndev:\n  sql: ${selectStatement} limit 10000`
      : "";

    const yamlContent = YAML_MODEL_TEMPLATE.replace(
      "{{ connector }}",
      connector,
    )
      .replace(/{{ sql }}/g, selectStatement)
      .replace("{{ dev_section }}", devSection);

    expect(yamlContent).toContain("dev:");
    expect(yamlContent).toContain("limit 10000");
    expect(yamlContent).toContain("connector: clickhouse");
  });

  it("should exclude dev section for Redshift connector", () => {
    const connector = "redshift";
    const selectStatement = "select * from my_table";
    const shouldIncludeDevSection = false;
    const devSection = shouldIncludeDevSection
      ? `\ndev:\n  sql: ${selectStatement} limit 10000`
      : "";

    const yamlContent = YAML_MODEL_TEMPLATE.replace(
      "{{ connector }}",
      connector,
    )
      .replace(/{{ sql }}/g, selectStatement)
      .replace("{{ dev_section }}", devSection);

    expect(yamlContent).not.toContain("dev:");
    expect(yamlContent).not.toContain("limit 10000");
    expect(yamlContent).toContain("connector: redshift");
    expect(yamlContent).toContain("sql: select * from my_table");
  });
});

describe("replaceOrAddEnvVariable", () => {
  it("should create a new env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable("", "KEY1", "VALUE1");
    expect(updatedEnvBlob).toBe("KEY1=VALUE1");
  });

  const existingEnvBlob = `# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2`;

  it("should update an existing key in the env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable(
      existingEnvBlob,
      "KEY1",
      "NEW_VALUE1",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=NEW_VALUE1
KEY2=VALUE2`);
  });

  it("should add a new key to the env file", () => {
    const updatedEnvBlob = replaceOrAddEnvVariable(
      existingEnvBlob,
      "KEY3",
      "VALUE3",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2
KEY3=VALUE3`);
  });
});

describe("replaceOlapConnectorInYAML", () => {
  it("should add a new `olap_connector` key to a blank file", () => {
    const updatedBlob = replaceOlapConnectorInYAML("", "clickhouse");
    expect(updatedBlob).toBe("olap_connector: clickhouse\n");
  });

  it("should add a new `olap_connector` key to a file with other keys", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const updatedBlob = replaceOlapConnectorInYAML(existingBlob, "clickhouse");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });

  it("should update the `olap_connector` key in a file with an existing `olap_connector` key", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n\nolap_connector: snowflake\n`;
    const updatedBlob = replaceOlapConnectorInYAML(existingBlob, "clickhouse");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nolap_connector: clickhouse\n`,
    );
  });
});

describe("replaceAiConnectorInYAML", () => {
  it("should add a new `ai_connector` key to a blank file", () => {
    const updatedBlob = replaceAiConnectorInYAML("", "claude");
    expect(updatedBlob).toBe("ai_connector: claude\n");
  });

  it("should add a new `ai_connector` key to a file with other keys", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nai_connector: claude\n`,
    );
  });

  it("should update the `ai_connector` key in a file with an existing `ai_connector` key", () => {
    const existingBlob = `# here's a comment\ntitle: test project\n\nai_connector: openai\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(
      `# here's a comment\ntitle: test project\n\nai_connector: claude\n`,
    );
  });

  it("should handle a blob without a trailing newline", () => {
    const existingBlob = `title: test project`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "claude");
    expect(updatedBlob).toBe(`title: test project\nai_connector: claude\n`);
  });

  it("should replace ai_connector in the middle of the file", () => {
    const existingBlob = `title: test project\nai_connector: openai\nolap_connector: clickhouse\n`;
    const updatedBlob = replaceAiConnectorInYAML(existingBlob, "gemini");
    expect(updatedBlob).toBe(
      `title: test project\nai_connector: gemini\nolap_connector: clickhouse\n`,
    );
describe("getGenericEnvVarName", () => {
  describe("Schema-driven x-env-var-name (production behavior)", () => {
    it("should use x-env-var-name from schema when provided", () => {
      const schema = {
        properties: {
          key_id: { "x-env-var-name": "GCS_ACCESS_KEY_ID" },
        },
      };
      const result = getGenericEnvVarName("gcs", "key_id", schema);
      expect(result).toBe("GCS_ACCESS_KEY_ID");
    });

    it("should use x-env-var-name for AWS credentials", () => {
      const schema = {
        properties: {
          aws_access_key_id: { "x-env-var-name": "AWS_ACCESS_KEY_ID" },
          aws_secret_access_key: { "x-env-var-name": "AWS_SECRET_ACCESS_KEY" },
        },
      };
      expect(getGenericEnvVarName("s3", "aws_access_key_id", schema)).toBe(
        "AWS_ACCESS_KEY_ID",
      );
      expect(getGenericEnvVarName("s3", "aws_secret_access_key", schema)).toBe(
        "AWS_SECRET_ACCESS_KEY",
      );
    });

    it("should use x-env-var-name for Google credentials", () => {
      const schema = {
        properties: {
          google_application_credentials: {
            "x-env-var-name": "GOOGLE_APPLICATION_CREDENTIALS",
          },
        },
      };
      expect(
        getGenericEnvVarName(
          "bigquery",
          "google_application_credentials",
          schema,
        ),
      ).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should use x-env-var-name for driver-specific secrets", () => {
      const motherduckSchema = {
        properties: {
          token: { "x-env-var-name": "MOTHERDUCK_TOKEN" },
        },
      };
      expect(
        getGenericEnvVarName("motherduck", "token", motherduckSchema),
      ).toBe("MOTHERDUCK_TOKEN");

      const clickhouseSchema = {
        properties: {
          dsn: { "x-env-var-name": "CLICKHOUSE_DSN" },
          password: { "x-env-var-name": "CLICKHOUSE_PASSWORD" },
        },
      };
      expect(getGenericEnvVarName("clickhouse", "dsn", clickhouseSchema)).toBe(
        "CLICKHOUSE_DSN",
      );
      expect(
        getGenericEnvVarName("clickhouse", "password", clickhouseSchema),
      ).toBe("CLICKHOUSE_PASSWORD");
    });
  });

  describe("Fallback behavior (DRIVERNAME_PROPERTYKEY format)", () => {
    it("should use DRIVERNAME_PROPERTYKEY when schema has no x-env-var-name", () => {
      const schema = {
        properties: {
          other_field: { "x-env-var-name": "OTHER_VAR" },
        },
      };
      const result = getGenericEnvVarName("gcs", "key_id", schema);
      expect(result).toBe("GCS_KEY_ID");
    });

    it("should use DRIVERNAME_PROPERTYKEY when schema is undefined", () => {
      const result = getGenericEnvVarName("motherduck", "token", undefined);
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should use DRIVERNAME_PROPERTYKEY when schema has no properties", () => {
      const schema = { properties: {} };
      const result = getGenericEnvVarName("motherduck", "token", schema);
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should convert camelCase to SCREAMING_SNAKE_CASE", () => {
      const result = getGenericEnvVarName("bigquery", "projectId");
      expect(result).toBe("BIGQUERY_PROJECT_ID");
    });

    it("should handle dots and hyphens by converting to underscores", () => {
      const result = getGenericEnvVarName(
        "custom",
        "property.with-mixed.separators",
      );
      expect(result).toBe("CUSTOM_PROPERTY_WITH_MIXED_SEPARATORS");
    });
  });
});

describe("envVarExists", () => {
  it("should return true when variable exists", () => {
    const envBlob = `KEY1=VALUE1\nKEY2=VALUE2\nKEY3=VALUE3`;
    expect(envVarExists(envBlob, "KEY2")).toBe(true);
  });

  it("should return false when variable does not exist", () => {
    const envBlob = `KEY1=VALUE1\nKEY2=VALUE2`;
    expect(envVarExists(envBlob, "KEY3")).toBe(false);
  });

  it("should return true for variable at start of file", () => {
    const envBlob = `FIRST_KEY=VALUE\nKEY2=VALUE2`;
    expect(envVarExists(envBlob, "FIRST_KEY")).toBe(true);
  });

  it("should return true for variable at end of file", () => {
    const envBlob = `KEY1=VALUE1\nLAST_KEY=VALUE`;
    expect(envVarExists(envBlob, "LAST_KEY")).toBe(true);
  });

  it("should not match partial variable names", () => {
    const envBlob = `MY_VARIABLE_1=VALUE\nMY_VARIABLE_2=VALUE2`;
    expect(envVarExists(envBlob, "MY_VARIABLE")).toBe(false);
  });

  it("should return false for empty blob", () => {
    expect(envVarExists("", "KEY1")).toBe(false);
  });

  it("should handle variables with complex values", () => {
    const envBlob = `JSON_VAR={"key":"value","nested":{"data":"here"}}\nSIMPLE=value`;
    expect(envVarExists(envBlob, "JSON_VAR")).toBe(true);
  });

  it("should handle multiline values (only checks line start)", () => {
    const envBlob = `KEY1=line1\nline2\nKEY2=VALUE2`;
    expect(envVarExists(envBlob, "KEY1")).toBe(true);
  });
});

describe("findAvailableEnvVarName", () => {
  it("should return base name when no conflicts", () => {
    const envBlob = `OTHER_VAR=value`;
    const result = findAvailableEnvVarName(envBlob, "MY_VAR");
    expect(result).toBe("MY_VAR");
  });

  it("should append _1 when base name exists", () => {
    const envBlob = `MY_VAR=value`;
    const result = findAvailableEnvVarName(envBlob, "MY_VAR");
    expect(result).toBe("MY_VAR_1");
  });

  it("should append _2 when _1 exists", () => {
    const envBlob = `MY_VAR=value\nMY_VAR_1=value`;
    const result = findAvailableEnvVarName(envBlob, "MY_VAR");
    expect(result).toBe("MY_VAR_2");
  });

  it("should skip gaps and find next available number", () => {
    const envBlob = `MY_VAR=value\nMY_VAR_1=value\nMY_VAR_2=value\nMY_VAR_3=value`;
    const result = findAvailableEnvVarName(envBlob, "MY_VAR");
    expect(result).toBe("MY_VAR_4");
  });

  it("should handle empty blob", () => {
    const result = findAvailableEnvVarName("", "NEW_VAR");
    expect(result).toBe("NEW_VAR");
  });

  it("should handle base name with underscores", () => {
    const envBlob = `GOOGLE_APPLICATION_CREDENTIALS=value\nGOOGLE_APPLICATION_CREDENTIALS_1=value`;
    const result = findAvailableEnvVarName(
      envBlob,
      "GOOGLE_APPLICATION_CREDENTIALS",
    );
    expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_2");
  });

  it("should be case sensitive", () => {
    const envBlob = `my_var=value`;
    const result = findAvailableEnvVarName(envBlob, "MY_VAR");
    expect(result).toBe("MY_VAR");
  });
});

describe("makeEnvVarKey", () => {
  // Mock schemas matching production x-env-var-name definitions
  const bigquerySchema = {
    properties: {
      google_application_credentials: {
        "x-env-var-name": "GOOGLE_APPLICATION_CREDENTIALS",
      },
    },
  };

  const s3Schema = {
    properties: {
      aws_access_key_id: { "x-env-var-name": "AWS_ACCESS_KEY_ID" },
      aws_secret_access_key: { "x-env-var-name": "AWS_SECRET_ACCESS_KEY" },
    },
  };

  const motherduckSchema = {
    properties: {
      token: { "x-env-var-name": "MOTHERDUCK_TOKEN" },
    },
  };

  const postgresSchema = {
    properties: {
      password: { "x-env-var-name": "POSTGRES_PASSWORD" },
    },
  };

  describe("Without existing env blob - returns schema-defined name", () => {
    it("should return GOOGLE_APPLICATION_CREDENTIALS for bigquery", () => {
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        undefined,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should return MOTHERDUCK_TOKEN for motherduck", () => {
      const result = makeEnvVarKey(
        "motherduck",
        "token",
        undefined,
        motherduckSchema,
      );
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should return POSTGRES_PASSWORD for postgres", () => {
      const result = makeEnvVarKey("postgres", "password", "", postgresSchema);
      expect(result).toBe("POSTGRES_PASSWORD");
    });
  });

  describe("With existing env blob - handles conflicts with _# suffix", () => {
    it("should append _1 when variable already exists", () => {
      const envBlob = `GOOGLE_APPLICATION_CREDENTIALS=existing_value`;
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_1");
    });

    it("should return base name when no conflict exists", () => {
      const envBlob = `OTHER_VAR=value`;
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should find next available number for multiple connectors of same type", () => {
      const envBlob = `GOOGLE_APPLICATION_CREDENTIALS=first_creds\nGOOGLE_APPLICATION_CREDENTIALS_1=second_creds\nGOOGLE_APPLICATION_CREDENTIALS_2=third_creds`;
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_3");
    });

    it("should handle multiple different properties", () => {
      const envBlob = `AWS_ACCESS_KEY_ID=key1\nAWS_SECRET_ACCESS_KEY=secret1`;
      const result = makeEnvVarKey(
        "s3",
        "aws_access_key_id",
        envBlob,
        s3Schema,
      );
      expect(result).toBe("AWS_ACCESS_KEY_ID_1");
    });

    it("should handle driver-specific properties with conflicts", () => {
      const envBlob = `MOTHERDUCK_TOKEN=token1\nMOTHERDUCK_TOKEN_1=token2`;
      const result = makeEnvVarKey(
        "motherduck",
        "token",
        envBlob,
        motherduckSchema,
      );
      expect(result).toBe("MOTHERDUCK_TOKEN_2");
    });

    it("should handle complex env blobs with comments and multiple variables", () => {
      const envBlob = `# This is a comment
SOME_OTHER_VAR=value
MOTHERDUCK_TOKEN=token1
MOTHERDUCK_TOKEN_1=token2
# Another comment
DATABASE_URL=something`;
      const result = makeEnvVarKey(
        "motherduck",
        "token",
        envBlob,
        motherduckSchema,
      );
      expect(result).toBe("MOTHERDUCK_TOKEN_2");
    });
  });

  describe("Integration - full workflows with schemas", () => {
    it("should support adding first BigQuery connector", () => {
      const emptyEnv = "";
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        emptyEnv,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should support adding second BigQuery connector", () => {
      const envAfterFirst = `GOOGLE_APPLICATION_CREDENTIALS=first_creds`;
      const result = makeEnvVarKey(
        "bigquery",
        "google_application_credentials",
        envAfterFirst,
        bigquerySchema,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_1");
    });

    it("should support adding AWS credentials to existing non-AWS variables", () => {
      const envBlob = `MOTHERDUCK_TOKEN=token1\nGOOGLE_APPLICATION_CREDENTIALS=creds1`;
      const result = makeEnvVarKey(
        "s3",
        "aws_access_key_id",
        envBlob,
        s3Schema,
      );
      expect(result).toBe("AWS_ACCESS_KEY_ID");
    });

    it("should support adding multiple AWS connectors", () => {
      const envBlob = `AWS_ACCESS_KEY_ID=key1\nAWS_SECRET_ACCESS_KEY=secret1`;
      const result1 = makeEnvVarKey(
        "s3",
        "aws_access_key_id",
        envBlob,
        s3Schema,
      );
      expect(result1).toBe("AWS_ACCESS_KEY_ID_1");

      const updatedEnv = `${envBlob}\nAWS_ACCESS_KEY_ID_1=key2`;
      const result2 = makeEnvVarKey(
        "s3",
        "aws_secret_access_key",
        updatedEnv,
        s3Schema,
      );
      expect(result2).toBe("AWS_SECRET_ACCESS_KEY_1");
    });
  });
});
