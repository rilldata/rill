package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;

import java.sql.SQLException;

public class SqlConverter
{
  private CalciteToolbox calciteToolbox;

  public SqlConverter(String schema) throws SQLException
  {
    calciteToolbox = new CalciteToolbox(new StaticSchemaProvider(schema),
                                        PostgresqlSqlDialect.DEFAULT,
                                        null
    );
  }

  public String convert(String sql)
  {
    try {
      return calciteToolbox.getRunnableQuery(sql);
    }
    catch (Exception e) {
      e.printStackTrace();
      return null;
    }
  }
}
