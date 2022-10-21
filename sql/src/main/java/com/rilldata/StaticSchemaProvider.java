package com.rilldata;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.rilldata.calcite.JsonSchema;
import com.rilldata.calcite.StaticSchema;
import org.apache.calcite.jdbc.CalciteSchema;
import org.apache.calcite.schema.SchemaPlus;

import java.io.IOException;
import java.util.List;
import java.util.function.Supplier;

public class StaticSchemaProvider implements Supplier<SchemaPlus>
{
  private final List<StaticSchema> staticSchemas;

  public StaticSchemaProvider(String jsonSchema) throws IOException
  {
    staticSchemas = List.of(new StaticSchema(new ObjectMapper().readValue(jsonSchema, JsonSchema.class)));
  }

  public StaticSchemaProvider(List<JsonSchema> jsonSchemas)
  {
    staticSchemas = List.copyOf(jsonSchemas.stream().map(StaticSchema::new).toList());
  }

  @Override
  public SchemaPlus get()
  {
    SchemaPlus rootSchema = CalciteSchema.createRootSchema(false).plus();
    for (StaticSchema staticSchema : staticSchemas) {
      rootSchema.add(staticSchema.getName(), staticSchema);
    }
    return rootSchema;
  }
}
