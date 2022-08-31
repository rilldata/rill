package com.rilldata.calcite;

import org.apache.calcite.rel.type.RelDataType;
import org.apache.calcite.rel.type.RelDataTypeFactory;
import org.apache.calcite.schema.impl.AbstractTable;
import org.apache.calcite.sql.type.SqlTypeName;

import java.util.List;
import java.util.Locale;

public class StaticTable extends AbstractTable
{
  private final List<JsonColumn> columns;

  public StaticTable(JsonTable table)
  {
    columns = table.columns;
  }

  @Override
  public RelDataType getRowType(RelDataTypeFactory typeFactory)
  {
    RelDataTypeFactory.FieldInfoBuilder builder = typeFactory.builder();
    for (JsonColumn column : columns) {
      builder.add(column.name, typeFactory.createSqlType(fromDuckDb(column.type)));
    }
    return builder.build();
  }

  public SqlTypeName fromDuckDb(String type)
  {
    String s = type.toUpperCase(Locale.ENGLISH);
    switch (s) {
      case "HUGEINT":
        return SqlTypeName.DECIMAL ;
      case "BLOB":
          return SqlTypeName.BINARY;
      case "UBIGINT":
        return SqlTypeName.BIGINT;
      case "UINT":
        return SqlTypeName.INTEGER;
      case "USMALLINT":
        return SqlTypeName.SMALLINT;
      case "UTINYINT":
        return  SqlTypeName.TINYINT;
      case "TIMESTAMP WITH TIME ZONE":
        return SqlTypeName.TIMESTAMP_WITH_LOCAL_TIME_ZONE;
      case "LIST":
        return SqlTypeName.ARRAY;
      case "STRUCT":
        return SqlTypeName.STRUCTURED;
      case "UUID":
        return SqlTypeName.VARCHAR;
      default:
        return SqlTypeName.valueOf(s);
    }
  }
}
