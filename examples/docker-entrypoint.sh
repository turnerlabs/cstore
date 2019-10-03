#!/bin/sh

#################################################################
# Use a version of cstore compatible with the developer's os to 
# run locally. 
#
# -l                 = converts the output to be log friendly 
# -e                 = sends the raw configuration to Stdout
# -g terminal-export = sends the export commands to Stdout 
# -t $CONFIG_ENV     = defines which enivironment to restore
# -v $CONFIG_VER"    = defines which version to restore
#
#################################################################
echo "Loading configuration for $CONFIG_ENV $CONFIG_VER."

#-----------------------------------------------------------------
# OPTION ONE: RESTORE ENV VARS
#-----------------------------------------------------------------
eval $(cstore pull -l -g terminal-export -t $CONFIG_ENV -v $CONFIG_VER) > /dev/null
if [ -z "$ANY_ENV_VAR_PULLED_BY_CSTORE" ]; then
    echo "level=fatal: cstore failed to get environment."
    exit 1
fi

#-----------------------------------------------------------------
# OPTION TWO: RESTORE JSON FILE 
#-----------------------------------------------------------------
# cstore pull -le -t $CONFIG_ENV -v $CONFIG_VER > config.json
# if grep -q $ANY_TEXT_PULLED_BY_CSTORE 'config.json'; then
#     echo "level=fatal: cstore failed to get environment."
#     exit 1
# fi

#################################################################
# 'exec' is a functionality of an operating system that runs an 
# executable file in the context of an already existing process, 
# replacing the previous executable. This allows the app to run
# as PID 1 to receive OS signals.
#
# Run 'docker exec {CONTAINER} ps' to view processes.
#################################################################
exec ./app