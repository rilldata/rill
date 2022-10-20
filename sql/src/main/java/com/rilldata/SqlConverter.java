package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.tools.ValidationException;

import java.io.IOException;

public class SqlConverter
{
  private final CalciteToolbox calciteToolbox;

  public SqlConverter(String catalog) throws IOException
  {
    calciteToolbox = CalciteToolbox.buildToolbox(catalog);
  }

  public String convert(String sql, SqlDialect sqlDialect) throws ValidationException, SqlParseException
  {
    return calciteToolbox.getRunnableQuery(sql, sqlDialect);
  }

  public byte[] getAST(String sql)
  {
    try {
      return calciteToolbox.getAST(sql, false);
    } catch (Exception e) {
      e.printStackTrace(); // todo level-logging for native libraries?
      // in case of error returning an AST containing StringLiteral with error messages as the top most node
      return calciteToolbox.getAST(
          SqlLiteral.createCharString(String.format("{'error': '%s'}", e.getMessage()), new SqlParserPos(0, 0)));
    }
  }
}
