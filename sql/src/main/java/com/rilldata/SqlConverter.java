package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.models.SqlCreateMetricsView;
import com.rilldata.calcite.models.SqlCreateSource;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.tools.ValidationException;

public class SqlConverter
{
  private CalciteToolbox calciteToolbox;

  public SqlConverter(String schema)
  {
    calciteToolbox = new CalciteToolbox(new StaticSchemaProvider(schema), null);
  }

  public String convert(String sql, SqlDialect sqlDialect) throws ValidationException, SqlParseException
  {
    return calciteToolbox.getRunnableQuery(sql, sqlDialect);
  }

  public byte[] createSource(String sourceDef)
  {
    try {
      SqlCreateSource sqlCreateSource = calciteToolbox.createSource(sourceDef);
      return calciteToolbox.getAST(sqlCreateSource);
    } catch (Exception e) {
      e.printStackTrace(); // todo level-logging for native libraries?
      // in case of error returning an AST containing StringLiteral with error messages as the top most node
      return calciteToolbox.getAST(
          SqlLiteral.createCharString(String.format("{'error': '%s'}", e.getMessage()), new SqlParserPos(0, 0)));
    }
  }

  public byte[] createMetricsView(String metricsViewDef)
  {
    try {
      SqlCreateMetricsView sqlCreateMetricsView = calciteToolbox.createMetricsView(metricsViewDef);
      return calciteToolbox.getAST(sqlCreateMetricsView);
    } catch (Exception e) {
      e.printStackTrace(); // todo level-logging for native libraries?
      // in case of error returning an AST containing StringLiteral with error messages as the top most node
      return calciteToolbox.getAST(
          SqlLiteral.createCharString(String.format("{'error': '%s'}", e.getMessage()), new SqlParserPos(0, 0)));
    }
  }
}
