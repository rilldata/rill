package com.rilldata.calcite;

import org.apache.calcite.schema.Table;
import org.apache.calcite.schema.impl.AbstractSchema;

import java.util.HashMap;
import java.util.Map;

public class StaticSchema extends AbstractSchema
{
  Map<String, Table> tables = new HashMap<>();

  public StaticSchema(JsonSchema schema) {
    for (JsonDbEntity table : schema.entities) {
      tables.put(table.name, new StaticTable(table));
    }
  }

  @Override
  protected Map<String, Table> getTableMap()
  {
    return tables;
  }
}
