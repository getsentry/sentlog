#!/bin/sh

# For compatibility with older entrypoints
if [ "${1}" == "sentlog" ]; then
  shift
elif [ "${1}" == "sh" ] || [ "${1}" == "/bin/sh" ]; then
  exec "$@"
fi

exec /bin/sentlog "$@"