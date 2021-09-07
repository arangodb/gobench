#!/usr/bin/fish

set proto $argv[1]

function do
  set parallelism $argv[1]
  set connections $argv[2]
  echo "Parallelism $parallelism, Connections $connections"
  date
  ./gobench2 --auth.user=root --auth.pass="" --endpoint=https://127.0.0.1:8529 \
             --nrConnections="$connections" --nrRequests=200000 \
             --parallelism="$parallelism" --protocol="$proto" --useTLS=true
  date
  echo
end

do 1 1
do 2 2
do 4 4
do 8 8
do 16 1
do 16 4
do 16 16
do 128 128

