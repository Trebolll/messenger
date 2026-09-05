# syntax=docker/dockerfile:1
FROM maven:3.9.11-eclipse-temurin-26 AS build
WORKDIR /workspace
COPY <<EOF pom.xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>ru.lambdahub</groupId>
    <artifactId>lambdahub-workspace</artifactId>
    <version>0.1.0-SNAPSHOT</version>
    <packaging>pom</packaging>
    <modules>
        <module>common-security</module>
        <module>common-events</module>
        <module>common-web</module>
        <module>common-kafka-starter</module>
        <module>validation-vavr-starter</module>
        <module>media-service</module>
    </modules>
</project>
EOF
COPY common-security common-security
COPY common-events common-events
COPY common-web common-web
COPY common-kafka-starter common-kafka-starter
COPY validation-vavr-starter validation-vavr-starter
COPY media-service media-service
RUN mvn -pl media-service/media-service-impl -am -DskipTests package

FROM eclipse-temurin:26-jre
WORKDIR /app
COPY --from=build /workspace/media-service/media-service-impl/target/media-service-impl-*.jar app.jar
EXPOSE 8084
ENTRYPOINT ["java","-jar","/app/app.jar"]
