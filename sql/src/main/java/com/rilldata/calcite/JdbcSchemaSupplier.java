package com.rilldata.calcite;

import org.apache.calcite.adapter.jdbc.JdbcSchema;
import org.apache.calcite.jdbc.CalciteSchema;
import org.apache.calcite.schema.SchemaPlus;

import javax.sql.DataSource;
import java.util.List;
import java.util.function.Supplier;

public class JdbcSchemaSupplier implements Supplier<SchemaPlus>
{
  private final DataSource dataSource;
  private List<String> datasourceSchemaNames;

  public JdbcSchemaSupplier(DataSource dataSource, List<String> schemaNames) {
    this.datasourceSchemaNames = schemaNames;
    this.dataSource = dataSource;
  }

  public DataSource getDataSource()
  {
    return dataSource;
  }

  @Override
  public SchemaPlus get()
  {
    final SchemaPlus rootSchema = CalciteSchema.createRootSchema(false).plus();
    /*
     TODO once these features are released - https://github.com/duckdb/duckdb/issues/3906
      consider using org.apache.calcite.adapter.jdbc.JdbcCatalogSchema
     */
    for (String datasourceSchema : datasourceSchemaNames) {
      JdbcSchema jdbcSchema = JdbcSchema.create(rootSchema, datasourceSchema, dataSource, null, datasourceSchema);
      rootSchema.add(datasourceSchema, jdbcSchema);
    }
    return rootSchema;
  }
}
