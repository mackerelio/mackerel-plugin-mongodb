#!/bin/sh

prog=$(basename $0)
if ! [ -S /var/run/docker.sock ]
then
	echo "$prog: there are no running docker" >&2
	exit 2
fi

cd $(dirname $0)
PATH=$(pwd):$PATH
plugin=$(basename $(pwd))
if ! which $plugin >/dev/null
then
	echo "$prog: $plugin is not installed" >&2
	exit 2
fi

user=root
password=passpass
port=27017

for v in 6.0 5.0 4.4 4.2 4.0 3.6
do
  docker run -d \
	--name test-$plugin \
	-p $port:$port \
	-e MONGO_INITDB_ROOT_USERNAME=$user \
	-e MONGO_INITDB_ROOT_PASSWORD=$password \
	mongo:$v
  trap 'docker stop test-$plugin; docker rm test-$plugin; exit 1' 1 2 3 15
  sleep 10

  if $plugin -port $port -username=$user -password $password
  then
    echo OK: $v
  else
    status=$?
    echo NG: $v
  fi

  docker stop "test-$plugin"
  docker rm "test-$plugin"
done
exit $status
