#!/bin/sh

#Verify Mode
if [[ "remote" == "${SSH_MODE}" ]]; then
  /whoami &
  while [ "$(nc -z $SSH_LOCAL_HOST $SSH_LOCAL_PORT </dev/null; echo $?)" !=  "0" ];
    do sleep 5;
    echo "Waiting for LOCAL PORT is up and responding";
  done
else
  echo "Waiting...";
  sleep 10;
fi

sleep 5;
/app/ssh-client