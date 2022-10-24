package com.rilldata;

import com.google.protobuf.InvalidProtocolBufferException;
import com.rilldata.calcite.dialects.Dialects;
import org.graalvm.nativeimage.IsolateThread;
import org.graalvm.nativeimage.c.function.CEntryPoint;
import org.graalvm.nativeimage.c.function.CFunctionPointer;
import org.graalvm.nativeimage.c.function.InvokeCFunctionPointer;
import org.graalvm.nativeimage.c.type.CCharPointer;
import org.graalvm.nativeimage.c.type.CIntPointer;
import org.graalvm.nativeimage.c.type.CTypeConversion;
import org.graalvm.nativeimage.c.type.VoidPointer;
import org.graalvm.word.WordFactory;
import com.rilldata.protobuf.generated.Requests;

import java.nio.ByteBuffer;
import java.util.Base64;

/**
 * This class contains an entry point (a function callable from a native executable, ie C/Go executable).
 */
public class SqlConverterEntrypoint
{
//  private static SqlConverter sqlConverter;

  interface AllocatorFn extends CFunctionPointer
  {
    @InvokeCFunctionPointer
    CCharPointer call(long size);
  }

  public static Requests.Response transpile(Requests.Request r) {
    Requests.TranspileRequest transpileRequest = r.getTranspileRequest();
    String sql = transpileRequest.getSql();
    Requests.Dialect dialect = transpileRequest.getDialect();
    try {
      SqlConverter sqlConverter = new SqlConverter(transpileRequest.getSchema());
      String transpiledSql = sqlConverter.convert(sql, Dialects.valueOf(dialect.name()).getSqlDialect());
      return Requests.Response
          .newBuilder()
          .setTranspileResponse(Requests.TranspileResponse.newBuilder().setSql(transpiledSql).build())
          .build();
    } catch (Exception e) {
      e.printStackTrace();
      return Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage(e.toString()).build())
          .build();
    }
  }

  @CEntryPoint(name = "processPbRequest")
  public static CCharPointer processPbRequest(
      IsolateThread thread,
      AllocatorFn allocatorFn,
      VoidPointer request,
      int inSize,
      CIntPointer outSize
  ) {
    ByteBuffer buf = CTypeConversion.asByteBuffer(request, inSize);
    byte[] arr = new byte[buf.limit()];
    buf.get(arr);
    byte[] out = processPbBytes(arr);
    CCharPointer charPointer = convertToCCharPointer2(allocatorFn, out);
    outSize.write(out.length);
    return charPointer;
  }

  public static byte[] processPbBytes(byte[] in) {
    try {
      Requests.Request r = Requests.Request.parseFrom(in);
      if (r.hasParseRequest()) {
        Requests.ParseRequest parseRequest = r.getParseRequest();
        String sql = parseRequest.getSql();
        SqlConverter sqlConverter = new SqlConverter(parseRequest.getSchema());
        return sqlConverter.getAST(sql);
      } else if (r.hasTranspileRequest()) {
        return transpile(r).toByteArray();
      }
      return Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage("Empty request").build())
          .build().toByteArray();
    }
    catch (InvalidProtocolBufferException e) {
      e.printStackTrace();
      return Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage("Invalid request" + e.getMessage()).build())
          .build().toByteArray();
    }
  }

  @CEntryPoint(name = "processPbRequest")
  public static CCharPointer processPbRequest(
      IsolateThread thread,
      AllocatorFn allocatorFn,
      VoidPointer request,
      CIntPointer inSize,
      CIntPointer outSize
  ) {
    int inSz = inSize.read();
    System.out.println("inSize " + inSz);
    ByteBuffer buf = CTypeConversion.asByteBuffer(request, inSz);
    byte[] arr = new byte[buf.limit()];
    buf.get(arr);
    byte[] out = processPbBytes(arr);
    CCharPointer charPointer = convertToCCharPointerNoZero(allocatorFn, out);
    outSize.write(out.length);
    return charPointer;
  }

  @CEntryPoint(name = "request")
  public static CCharPointer processBase64Request(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer request)
  {
    String b64String = CTypeConversion.toJavaString(request);
    byte[] decoded = Base64.getDecoder().decode(b64String);
    byte[] out = processPbBytes(decoded);
    byte[] b64response = Base64.getEncoder().encode(out);
    return convertToCCharPointer(allocatorFn, new String(b64response));
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

  private static CCharPointer convertToCCharPointer2(AllocatorFn allocatorFn, byte[] b)
  {
    CCharPointer a = allocatorFn.call(b.length + 1);
    for (int i = 0; i < b.length; i++) {
      a.write(i, b[i]);
    }
    return a;
  }

  private static CCharPointer convertToCCharPointerNoZero(AllocatorFn allocatorFn, byte[] b)
  {
    CCharPointer a = allocatorFn.call(b.length + 1);
    for (int i = 0; i < b.length; i++) {
      a.write(i, b[i]);
    }
    return a;
  }
}
