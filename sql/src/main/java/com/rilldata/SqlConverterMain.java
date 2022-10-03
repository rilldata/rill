package com.rilldata;

import com.rilldata.calcite.dialects.Dialects;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.ValidationException;

import java.io.IOException;
import java.sql.SQLException;

/**
 * This class is used to create a native executable (instead of the native shared library).
 * The native executable can be used to test GraalVM output without compiling Go native executable.
 **/
public class SqlConverterMain
{
  public static void main(String[] args) throws SQLException, IOException, ValidationException, SqlParseException
  {
    String s = new String(SqlConverterMain.class.getResourceAsStream("/schema.json").readAllBytes());
    SqlConverter sqlConverter = new SqlConverter(s);
    if (args.length == 1) {
      System.out.println(sqlConverter.convert(args[0], Dialects.DUCKDB.getSqlDialect()));
    } else if (args.length == 2) {
      Dialects dialectEnum = Dialects.valueOf(args[1].toUpperCase());
      System.out.println(sqlConverter.convert(args[0], dialectEnum.getSqlDialect()));
    } else {
      System.out.println(
          sqlConverter.convert("select \"name\" from \"main\".\"heroes\"", Dialects.DUCKDB.getSqlDialect()));
    }
  }
}
