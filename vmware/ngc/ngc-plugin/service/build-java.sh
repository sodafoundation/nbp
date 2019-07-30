#!/bin/sh
# Mac OS script starting an Ant build of the current flex project
# Note: if Ant runs out of memory try defining ANT_OPTS=-Xmx512M

if [ -z "$ANT_HOME" ] || [ ! -f "${ANT_HOME}"/bin/ant ]
then
   echo BUILD FAILED: You must set the environment variable ANT_HOME to your Apache Ant folder
   exit 1
fi

"${ANT_HOME}"/bin/ant -f build-java.xml

exit 0
