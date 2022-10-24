package com.rilldata;

import com.google.protobuf.InvalidProtocolBufferException;
import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.protobuf.generated.Requests;
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
    proceessProtobufTranspileRequest();
    processSchemaDependentSql();
  }

  private static void processSchemaDependentSql() throws IOException, ValidationException, SqlParseException
  {
    String s = new String(SqlConverterMain.class.getResourceAsStream("/schema.json").readAllBytes());
    SqlConverter sqlConverter = new SqlConverter(s);
    System.out.println(
        sqlConverter.convert("select \"name\" from \"main\".\"heroes\"", Dialects.DUCKDB.getSqlDialect()));
  }

  private static void proceessProtobufTranspileRequest() throws InvalidProtocolBufferException
  {
    Requests.Request request = Requests.Request
        .newBuilder()
        .setTranspileRequest(Requests.TranspileRequest
                                 .newBuilder()
                                 .setSql("select 1")
                                 .setDialect(Requests.Dialect.DUCKDB)
                                 .setSchema("""
                                 { 
                                  "tables": []
                                 }
                                 """)
                                 .build()

        )
        .build();
    byte[] bytes = SqlConverterEntrypoint.processPbBytes(request.toByteArray());
    Requests.Response response = Requests.Response.parseFrom(bytes);
    System.out.println(response.getTranspileResponse().getSql());
  }
}
