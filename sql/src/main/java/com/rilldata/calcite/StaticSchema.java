package com.rilldata.calcite;

import org.apache.calcite.schema.Table;
import org.apache.calcite.schema.impl.AbstractSchema;

import java.util.HashMap;
import java.util.Map;

public class StaticSchema extends AbstractSchema
{
  String name;
  Map<String, Table> tables = new HashMap<>();

  public StaticSchema(JsonSchema schema) {
    this.name = schema.name;
    for (JsonTable table : schema.tables) {
      tables.put(table.name, new StaticTable(table));
    }
  }

  public String getName() {
    return name;
  }

  @Override
  protected Map<String, Table> getTableMap()
  {
    return tables;
  }
}
