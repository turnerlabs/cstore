#!/bin/sh

#################################################################
# Use a local version of cstore compatible with the developer's
# os to test locally.
#
# eval $( cstore pull -e -t $CONFIG_ENV) 
#################################################################
echo "Loading configuration for $CONFIG_ENV."

eval $(./tools/config/cstore_linux_amd64 pull -c aws-sdk -e -t $CONFIG_ENV) > /dev/null

#################################################################
# 'exec' is a functionality of an operating system that runs an 
# executable file in the context of an already existing process, 
# replacing the previous executable. This allows the app to run
# as PID 1 to receive OS signals.
#
# 'docker exec {CONTAINER} ps'
#################################################################
exec ./my-application