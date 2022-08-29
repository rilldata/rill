package com.rilldata;

import com.rilldata.calcite.StaticSchema;
import com.rilldata.calcite.StaticSchemaFactory;
import org.apache.calcite.jdbc.CalciteSchema;
import org.apache.calcite.schema.SchemaPlus;

import java.util.function.Supplier;

class StaticSchemaProvider implements Supplier<SchemaPlus>
{
  private String schema;

  public StaticSchemaProvider(String schema)
  {
    this.schema = schema;
  }

  @Override
  public SchemaPlus get()
  {
    try {
      StaticSchema staticSchema = StaticSchemaFactory.create(schema);
      SchemaPlus rootSchema = CalciteSchema.createRootSchema(false).plus();

      rootSchema.add("main", staticSchema);
      return rootSchema;
    }
    catch (Exception e) {
      throw new RuntimeException(e);
    }
  }
}
