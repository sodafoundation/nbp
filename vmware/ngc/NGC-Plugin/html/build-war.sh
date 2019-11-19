#!/bin/sh
# Mac OS script starting an Ant build of the current flex project
# Note: if Ant runs out of memory try defining ANT_OPTS=-Xmx512M

if [ -z "$ANT_HOME" ] || [ ! -f "${ANT_HOME}"/bin/ant ]
then
   echo BUILD FAILED: You must set the environment variable ANT_HOME to your Apache Ant folder
   exit 1
fi

if [ -z "$FLEX_HOME" ] || [ ! -f "$FLEX_HOME"/bin/mxmlc ]
then
   echo BUILD FAILED: You must set the environment variable FLEX_HOME to your Flex SDK folder,
   echo for instance: FLEX_HOME=\'/Applications/Adobe Flash Builder 4.7/sdks/4.6.0\'
   exit 1
fi

if [ -z "$VSPHERE_SDK_HOME" ] || [ ! -f "${VSPHERE_SDK_HOME}"/libs/vsphere-client-lib.swc ]
then
   echo BUILD FAILED: You must set the environment variable VSPHERE_SDK_HOME to your vSphere Web Client SDK folder
   exit 1
fi

"${ANT_HOME}"/bin/ant -f build-war.xml

exit 0
