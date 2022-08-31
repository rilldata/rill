package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;

import javax.sql.DataSource;
import java.sql.SQLException;

public class SqlConverter
{
  private DataSource datasource;
  private CalciteToolbox calciteToolbox;

  public SqlConverter(String schema) throws SQLException
  {
    calciteToolbox = new CalciteToolbox(new StaticSchemaProvider(schema),
                                        PostgresqlSqlDialect.DEFAULT,
                                        null
    );
  }

  public void initialize(String ddl) throws SQLException
  {
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

  public String inferMigrations(String json)
  {
    try {
      return calciteToolbox.getRunnableQuery(json);
    }
    catch (Exception e) {
      e.printStackTrace();
      return null;
    }
  }
}
