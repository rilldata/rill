package com.rilldata;

import java.io.IOException;
import java.sql.SQLException;

/**
 * This class is used to create a native executable (instead of the native shared library).
 * The native executable can be used to test GraalVM output without compiling Go native executable.
 **/
public class SqlConverterMain
{
  public static void main(String[] args) throws SQLException, IOException
  {
    String s = new String(SqlConverterMain.class.getResourceAsStream("/schema.json").readAllBytes());
    SqlConverter sqlConverter = new SqlConverter(s);
    if (args.length > 0) {
      System.out.println(sqlConverter.convert(args[0]));
    } else {
      System.out.println(sqlConverter.convert("select \"name\" from \"main\".\"heroes\""));
    }
  }
}
