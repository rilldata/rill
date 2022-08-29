package com.rilldata.calcite;

import org.apache.calcite.schema.Table;
import org.apache.calcite.schema.impl.AbstractSchema;

import java.util.HashMap;
import java.util.Map;

public class StaticSchema extends AbstractSchema
{
  Map<String, Table> tables = new HashMap<>();

  public StaticSchema(JsonSchema schema) {
    for (JsonTable table : schema.tables) {
      tables.put(table.name, new StaticTable(table));
    }
  }

  @Override
  protected Map<String, Table> getTableMap()
  {
    return tables;
  }
}
