'use strict';
const shell = require('shelljs');

function pull(cstorePath, tags) {
  // Silent is set to true to prevent cstore's output from being sent to stdout
  // where it would be logged compromising secrets.
  //
  // Add "-i" to the pull command when injectig secrets into configuration from
  // Secrets Manager.
  var result = shell.exec('./' + cstorePath + ' pull -le -t "' + tags + '"', {silent:true})

  if (result.code != 0) {
    throw new Error(result.stderr)
  }

  return JSON.parse(result.stdout)
}

module.exports.pull = pull;