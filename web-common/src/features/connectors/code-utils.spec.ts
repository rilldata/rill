import { describe, expect, it } from "vitest";
import {
  replaceOlapConnectorInYAML,
  replaceOrAddEnvVariable,
  getGenericEnvVarName,
  envVarExists,
  findAvailableEnvVarName,
  makeDotEnvConnectorKey,
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

describe("getGenericEnvVarName", () => {
  describe("Generic properties - no driver prefix", () => {
    it("should return GOOGLE_APPLICATION_CREDENTIALS for google_application_credentials", () => {
      const result = getGenericEnvVarName(
        "bigquery",
        "google_application_credentials",
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should return AWS_ACCESS_KEY_ID for aws_access_key_id", () => {
      const result = getGenericEnvVarName("s3", "aws_access_key_id");
      expect(result).toBe("AWS_ACCESS_KEY_ID");
    });

    it("should return AWS_SECRET_ACCESS_KEY for aws_secret_access_key", () => {
      const result = getGenericEnvVarName("athena", "aws_secret_access_key");
      expect(result).toBe("AWS_SECRET_ACCESS_KEY");
    });

    it("should return AZURE_STORAGE_CONNECTION_STRING for azure_storage_connection_string", () => {
      const result = getGenericEnvVarName(
        "adx",
        "azure_storage_connection_string",
      );
      expect(result).toBe("AZURE_STORAGE_CONNECTION_STRING");
    });

    it("should handle generic properties with different drivers", () => {
      const result = getGenericEnvVarName("redshift", "aws_access_token");
      expect(result).toBe("AWS_ACCESS_TOKEN");
    });
  });

  describe("Driver-specific properties - with driver prefix", () => {
    it("should return MOTHERDUCK_TOKEN for motherduck driver", () => {
      const result = getGenericEnvVarName("motherduck", "token");
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should return SNOWFLAKE_ACCOUNT for snowflake driver", () => {
      const result = getGenericEnvVarName("snowflake", "account");
      expect(result).toBe("SNOWFLAKE_ACCOUNT");
    });

    it("should return CLICKHOUSE_DSN for clickhouse driver", () => {
      const result = getGenericEnvVarName("clickhouse", "dsn");
      expect(result).toBe("CLICKHOUSE_DSN");
    });
  });

  describe("Case conversion - camelCase to SCREAMING_SNAKE_CASE", () => {
    it("should convert camelCase property names", () => {
      const result = getGenericEnvVarName("bigquery", "projectId");
      expect(result).toBe("BIGQUERY_PROJECT_ID");
    });

    it("should convert multiple transitions in camelCase", () => {
      const result = getGenericEnvVarName("postgres", "sslCertificatePath");
      expect(result).toBe("POSTGRES_SSL_CERTIFICATE_PATH");
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

describe("makeDotEnvConnectorKey", () => {
  describe("Without existing env blob - returns base generic name", () => {
    it("should return generic name for google_application_credentials", () => {
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should return driver-prefixed name for motherduck token", () => {
      const result = makeDotEnvConnectorKey("motherduck", "token");
      expect(result).toBe("MOTHERDUCK_TOKEN");
    });

    it("should handle undefined existingEnvBlob", () => {
      const result = makeDotEnvConnectorKey("snowflake", "account", undefined);
      expect(result).toBe("SNOWFLAKE_ACCOUNT");
    });

    it("should handle null existingEnvBlob", () => {
      const result = makeDotEnvConnectorKey("postgres", "password", "");
      expect(result).toBe("POSTGRES_PASSWORD");
    });
  });

  describe("With existing env blob - handles conflicts", () => {
    it("should append _1 when variable already exists", () => {
      const envBlob = `GOOGLE_APPLICATION_CREDENTIALS=existing_value`;
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_1");
    });

    it("should return base name when no conflict exists", () => {
      const envBlob = `OTHER_VAR=value`;
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should find next available number for multiple connectors of same type", () => {
      const envBlob = `GOOGLE_APPLICATION_CREDENTIALS=first_creds\nGOOGLE_APPLICATION_CREDENTIALS_1=second_creds\nGOOGLE_APPLICATION_CREDENTIALS_2=third_creds`;
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
        envBlob,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_3");
    });

    it("should handle multiple different properties", () => {
      const envBlob = `AWS_ACCESS_KEY_ID=key1\nAWS_SECRET_ACCESS_KEY=secret1`;
      const result = makeDotEnvConnectorKey(
        "athena",
        "aws_access_key_id",
        envBlob,
      );
      expect(result).toBe("AWS_ACCESS_KEY_ID_1");
    });

    it("should handle driver-specific properties with conflicts", () => {
      const envBlob = `MOTHERDUCK_TOKEN=token1\nMOTHERDUCK_TOKEN_1=token2`;
      const result = makeDotEnvConnectorKey("motherduck", "token", envBlob);
      expect(result).toBe("MOTHERDUCK_TOKEN_2");
    });

    it("should handle complex env blobs with comments and multiple variables", () => {
      const envBlob = `# This is a comment
SOME_OTHER_VAR=value
MOTHERDUCK_TOKEN=token1
MOTHERDUCK_TOKEN_1=token2
# Another comment
DATABASE_URL=something`;
      const result = makeDotEnvConnectorKey("motherduck", "token", envBlob);
      expect(result).toBe("MOTHERDUCK_TOKEN_2");
    });
  });

  describe("Integration - full workflows", () => {
    it("should support adding first BigQuery connector", () => {
      const emptyEnv = "";
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
        emptyEnv,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS");
    });

    it("should support adding second BigQuery connector", () => {
      const envAfterFirst = `GOOGLE_APPLICATION_CREDENTIALS=first_creds`;
      const result = makeDotEnvConnectorKey(
        "bigquery",
        "google_application_credentials",
        envAfterFirst,
      );
      expect(result).toBe("GOOGLE_APPLICATION_CREDENTIALS_1");
    });

    it("should support adding AWS credentials to existing non-AWS variables", () => {
      const envBlob = `MOTHERDUCK_TOKEN=token1\nGOOGLE_APPLICATION_CREDENTIALS=creds1`;
      const result = makeDotEnvConnectorKey(
        "athena",
        "aws_access_key_id",
        envBlob,
      );
      expect(result).toBe("AWS_ACCESS_KEY_ID");
    });

    it("should support adding multiple AWS connectors", () => {
      const envBlob = `AWS_ACCESS_KEY_ID=key1\nAWS_SECRET_ACCESS_KEY=secret1`;
      const result1 = makeDotEnvConnectorKey(
        "s3",
        "aws_access_key_id",
        envBlob,
      );
      expect(result1).toBe("AWS_ACCESS_KEY_ID_1");

      const updatedEnv = `${envBlob}\nAWS_ACCESS_KEY_ID_1=key2`;
      const result2 = makeDotEnvConnectorKey(
        "athena",
        "aws_secret_access_key",
        updatedEnv,
      );
      expect(result2).toBe("AWS_SECRET_ACCESS_KEY_1");
    });
  });
});
