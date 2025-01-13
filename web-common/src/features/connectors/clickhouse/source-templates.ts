import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { CLICKHOUSE_SOURCE_CONNECTORS } from "../connector-availability";

type ClickhouseSourceConnector = (typeof CLICKHOUSE_SOURCE_CONNECTORS)[number];

export function compileClickhouseSourceConnectorFile(
  connector: Omit<V1ConnectorDriver, "name"> & {
    name: ClickhouseSourceConnector;
  },
  formValues: Record<string, unknown>,
) {
  switch (connector.name) {
    case "azure":
      return `select * from azure('${formValues.path as string}');`;
    case "gcs":
      return `select * from gcs('${formValues.path as string}');`;
    case "s3":
      s3(
        path,
        [aws_access_key_id, aws_secret_access_key][
          (format, [structure, [compression]])
        ],
      );
      return `select * from s3('${formValues.path as string}');`;
    case "local_file": // mapping.csv
      return `select * from mapping`;
    case "postgres":
      return `select * from postgresql('${formValues.host as string}:${formValues.port as string}', 
'${formValues.database as string}', 
'${formValues.table as string}', 
'${formValues.user as string}', 
'${formValues.password as string}'
);`;
    case "mysql":
      return `select * from mysql('${formValues.host as string}:${formValues.port as string}', '${formValues.database as string}', '${formValues.user as string}', '${formValues.password as string}');`;
    default:
      assertNever(connector.name);
    // default:
    //   throw new Error(`Unsupported connector: ${connector.name}`);
  }
}

function assertNever(x: never): never {
  throw new Error(`Unexpected connector: ${x}`);
}
