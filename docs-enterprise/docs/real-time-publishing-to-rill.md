---
title: "Connect Real-time Publishing to Rill"
slug: "real-time-publishing-to-rill"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

## Streaming Platform

Connectivity is done with standard Apache Kafka client configuration allowing for easy in while also being secure. 
## Naming Conventions

A unique client identifier is provided.  For example, a company named "Foo Industries" could be provided a client identifier of "foo".  All topics for this client will be prefixed with "foo-" and the provided credentials will allow for you to produce and consume from those topics. In addition, only consumer groups with the same prefix are allowed.  **--group foo-123** is valid while **--group foo123** is invalid.

You are not allowed to create topics or perform any administration of the topics. Topics are created based on your business use case.  In addition a topic **{{client}}-text** is created for your connection testing.  

A unique client identifier is provided as part of onboarding. 
## Client Configuration

Data is encrypted in-flight with TLS certificates. These certificates are currently self-signed and require the installation of a certificate of trust.  Authentication and Authorization are handled by a provided key/secret pair leveraging SASL Scram. If your secret is internally compromised, please contact Rill Data immediately to generate and issue a new secret or even a new complete key/secret upon your request.

#### Provided Information

* Bootstrap Servers - the host(s) used to discover the Apache Kafka Cluster by your client(s).
* SASL Scram Credentials  - key and secret
* Rill Data CA Certificate - provided both as a .pem as well as in a pre-packaged .jks file with a simple password. 

Both the Scram `key` and `secret` however, are to be treated as confidential and should not be shared.  The secret is salted and hashed (SHA-512) and leverages the standard Apache Kafka Scram security module.  Security patches and updates are closely monitored.

For additional security, you can request the credentials have ONLY producing authorization, which would prevent your team to be able to read anything published to RillData services with these credentials.

#### Bootstrap Servers
```json
--bootstrap-servers {{ hostname }}:{{ port }},{{ hostname }}:{{ port }}
```
#### Java clients
```json
security.protocol=SASL_SSL
sasl.mechanism=SCRAM-SHA-512
sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required \\
    username="{{ key }}" \\
    password="{{ secret }}";
ssl.truststore.location={{ path to your JKS file with RillData Kafka Truststore }}
ssl.truststore.password={{ password for your truststore }}
```
#### librdkafka clients
```json
security.protocol=SASL_SSL
sasl.mechanism=SCRAM-SHA-512
sasl.username={{ key }}
sasl.password={{ secret }}
ssl.ca.location={{ path to the RillData pem file }}
```

## Topics

Topics are created by RillData.  As indicated by naming conventions, they are prefixed with your client identifier. Beyond that, they can be named whatever makes sense to your publishing pipeline, as long as they follow Apache Kafka Topic naming conventions.  Topics are created by Rill Data.  Retention time and other settings are part of the onboarding process.

It is recommended to publish data with compression. `compression.type` of the topic will be set to `producer` meaning it will honor what is set.

It is also recommended to submit with `acks`=`all` and to ensure the client processes and handles accordingly.  Additional recommended Kafka producer settings are provided as part of the onboarding process.
## Testing Connectivity

If your client is written in Java, it is recommended to use the `kafka-console-` commands to validate connection, since the configuration would be identical. If your client is written in a non-Java language and leverages the librdkafka library, then use kafkacat `kcat` for validation of the credentials.

Leverage the appropriate configuration files as shown above in the client setup recommendations.

###Java

####Publishing
```json
kafka-console-producer \\
    --bootstrap-server {{ server }}:{{ port }},{{ server }}:{{ port }} \\
    --producer.config ./client.conf \\
    --topic {{ client }}-text
    
```
####Consuming
```json
kafka-console-consumer \\
    --bootstrap-server {{ server }}:{{ port }},{{ server }}:{{ port }} \\
    --consumer.config ./client.conf \\
    --from-beginning \\
    --group {{ client }}-000 \\
    --topic {{ client }}-text
```
### librdkafka

If your client leverages **librdkafka** library, the best tool for establishing connectivity is the **kcat** (formally kafkacat) command-line tool.  This took leverages the librdkafka library so configuration for this would more closely resemble the configuration of your client.

####publishing
```json
kcat -b {{ server }}:{{ port }},{{ server }}:{{ port }} \\
    -F ./client.conf \\
    -P \\
    -t {{ client }}-text
```
###consuming
```json
kcat -b {{ server }}:{{ port }},{{ server }}:{{ port }} \\
    -F ./client.conf \\
    -C \\
    -t {{ client }}-text
```


### Trouble-Shooting

#### SSL Handshake

Leverage the following when starting your Kafka client (KAFKA_OPTS for kafka-console-producer or kafka-console-consumer) to help uncover any connectivity issues.

**-Djavax.net.debug=ssl,handshake**