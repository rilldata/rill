package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.protobuf.generated.SqlNodeProto;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.ValidationException;

public class SqlConverter
{
  private final CalciteToolbox calciteToolbox;

  public SqlConverter(String schema)
  {
    calciteToolbox = new CalciteToolbox(new StaticSchemaProvider(schema), null);
  }

  public String convert(String sql, SqlDialect sqlDialect) throws ValidationException, SqlParseException
  {
    return calciteToolbox.getRunnableQuery(sql, sqlDialect);
  }

  public SqlNodeProto getAST(String sql) throws ValidationException, SqlParseException
  {
    return calciteToolbox.getAST(sql, false);
  }
}
