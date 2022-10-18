package com.rilldata;

import com.google.protobuf.InvalidProtocolBufferException;
import com.rilldata.calcite.dialects.Dialects;
import org.graalvm.nativeimage.IsolateThread;
import org.graalvm.nativeimage.c.function.CEntryPoint;
import org.graalvm.nativeimage.c.function.CFunctionPointer;
import org.graalvm.nativeimage.c.function.InvokeCFunctionPointer;
import org.graalvm.nativeimage.c.type.CCharPointer;
import org.graalvm.nativeimage.c.type.CTypeConversion;
import org.graalvm.word.WordFactory;
import sql.v1.Requests;

/**
 * This class contains an entry point (a function callable from a native executable, ie C/Go executable).
 */
public class SqlConverterEntrypoint
{
  private static SqlConverter sqlConverter;

  interface AllocatorFn extends CFunctionPointer
  {
    @InvokeCFunctionPointer
    CCharPointer call(long size);
  }

  @CEntryPoint(name = "request")
  public static CCharPointer processRequest(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer request) {
    String s = CTypeConversion.toJavaString(request);

    try {
      Requests.Request r = Requests.Request.parseFrom(s.getBytes());
      Requests.ParseRequest parseRequest = r.getParseRequest();
      if (parseRequest.isInitialized()) {
        String sql = parseRequest.getSql();
        System.out.println(sql);
//        String dialect = parseRequest.getDialect();
//        String result = sqlConverter.convert(sql, Dialects.valueOf(dialect));
        byte[] bytes = Requests.Response
            .newBuilder()
            .setParseResponse(Requests.ParseResponse.newBuilder()
                                  .setSql("SELECT 1")
                                                    .build())
            .build()
            .toByteArray();
        return convertToCCharPointer(allocatorFn, new String(bytes));
      }
      Requests.Response build = Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage("Empty request").build())
          .build();
      return convertToCCharPointer(allocatorFn, new String(build.toByteArray()));
    }
    catch (InvalidProtocolBufferException e) {
      e.printStackTrace();
      Requests.Response build = Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage("Invalid request" + e.getMessage()).build())
          .build();
      return convertToCCharPointer(allocatorFn, new String(build.toByteArray()));
    }
  }

  @CEntryPoint(name = "convert_sql")
  public static CCharPointer convertSql(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer sql,
      CCharPointer schema, CCharPointer dialect
  )
  {
    try {
      String dialectString = CTypeConversion.toJavaString(dialect);
      Dialects dialectEnum = Dialects.valueOf(dialectString.toUpperCase());
      String javaSchemaString = CTypeConversion.toJavaString(schema);
      SqlConverter sqlConverter = new SqlConverter(javaSchemaString);
      String javaSqlString = CTypeConversion.toJavaString(sql);
      String runnableQuery = sqlConverter.convert(javaSqlString, dialectEnum.getSqlDialect());
      if (runnableQuery == null) {
        return WordFactory.nullPointer();
      }
      return convertToCCharPointer(allocatorFn, runnableQuery);
    } catch (Exception e) {
      e.printStackTrace(); // todo level-logging for native libraries?
      return convertToCCharPointer(allocatorFn, String.format("{'error': '%s'}", e.getMessage()));
    }
  }

  @CEntryPoint(name = "get_ast")
  public static CCharPointer getAST(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer sql,
      CCharPointer schema
  )
  {
    try {
      String javaSchemaString = CTypeConversion.toJavaString(schema);
      SqlConverter sqlConverter = new SqlConverter(javaSchemaString);
      String sqlString = CTypeConversion.toJavaString(sql);
      byte[] ast = sqlConverter.getAST(sqlString);
      return convertToCCharPointer(allocatorFn, ast);
    } catch (Exception e) {
      e.printStackTrace();
      return WordFactory.nullPointer();
    }
  }

  private static CCharPointer convertToCCharPointer(AllocatorFn allocatorFn, String javaString)
  {
    return convertToCCharPointer(allocatorFn, javaString.getBytes());
  }

  private static CCharPointer convertToCCharPointer(AllocatorFn allocatorFn, byte[] b)
  {
    CCharPointer a = allocatorFn.call(b.length + 1);
    for (int i = 0; i < b.length; i++) {
      a.write(i, b[i]);
    }
    a.write(b.length, (byte) 0);
    return a;
  }
}
