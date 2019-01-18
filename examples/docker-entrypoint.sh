#!/bin/sh

#################################################################
# Use a version of cstore compatible with the developer's os to 
# run locally. The -l flag converts the output to be more log
# friendly instead of terminal friendly. The -e sends each row of
# of configuration to stdout with a prefix of 'export'. The -t
# defines which tagged enivironment configuration to restore.
#
# "-v $CONFIG_VER" is optional and will pull config by version.
# Do not use this flag config is not versioned.
#################################################################
echo "Loading configuration for $CONFIG_ENV $CONFIG_VER."

eval $(cstore pull -le -t $CONFIG_ENV -v $CONFIG_VER) > /dev/null

#################################################################
# 'exec' is a functionality of an operating system that runs an 
# executable file in the context of an already existing process, 
# replacing the previous executable. This allows the app to run
# as PID 1 to receive OS signals.
#
# 'docker exec {CONTAINER} ps'
#################################################################
exec ./my-application