(
  set -e
  mvn package
  mvn dependency:copy-dependencies
  java -agentlib:native-image-agent=config-merge-dir=graalcfg,experimental-class-define-support -cp target/classes:target/dependency/*:target/test-classes com.rilldata.SqlConverterMain
  java -agentlib:native-image-agent=config-merge-dir=graalcfg,experimental-class-define-support -cp target/classes:target/dependency/*:target/test-classes com.rilldata.SqlConverterMain error
)
