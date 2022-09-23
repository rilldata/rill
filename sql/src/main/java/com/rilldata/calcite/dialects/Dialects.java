package com.rilldata.calcite.dialects;

import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;

public enum Dialects
{
  DUCKDB(PostgresqlSqlDialect.DEFAULT),
  DRUID(DruidDialect.DEFAULT);

  private SqlDialect sqlDialect;

  Dialects(SqlDialect sqlDialect)
  {
    this.sqlDialect = sqlDialect;
  }

  public SqlDialect getSqlDialect()
  {
    return sqlDialect;
  }
}
