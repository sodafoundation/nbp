@echo off
REM --- Windows script starting an Ant build of the current Java project
REM --- (if Ant runs out of memory try defining ANT_OPTS=-Xmx512M)

@IF not defined ANT_HOME (
   @echo BUILD FAILED: You must set the env variable ANT_HOME to your Apache Ant folder
   goto end
)

@call "%ANT_HOME%\bin\ant" -f build-java.xml

:end
