package com.rilldata;

import org.apache.calcite.adapter.jdbc.JdbcSchema;
import org.apache.calcite.jdbc.CalciteSchema;
import org.apache.calcite.schema.SchemaPlus;

import javax.sql.DataSource;
import java.util.Map;
import java.util.function.Supplier;

public class HsqlDbSchemaSupplier implements Supplier<SchemaPlus>
{
  private final DataSource dataSource;
  private Map<String, String> datasourceSchemaNames;

  public HsqlDbSchemaSupplier(Map<String, String> schemaNames) {
    this.datasourceSchemaNames = schemaNames;
    dataSource = JdbcSchema.dataSource(
        "jdbc:hsqldb:mem:db", "org.hsqldb.jdbc.JDBCDriver", "SA", null
    );
  }

  public DataSource getDataSource()
  {
    return dataSource;
  }

  @Override
  public SchemaPlus get()
  {
    final SchemaPlus rootSchema = CalciteSchema.createRootSchema(false).plus();
    for (String datasourceSchema : datasourceSchemaNames.keySet()) {
      JdbcSchema jdbcSchema = JdbcSchema.create(rootSchema, datasourceSchema, dataSource, null, datasourceSchemaNames.get(datasourceSchema));
      rootSchema.add(datasourceSchema, jdbcSchema);
    }
    return rootSchema;
  }
}
