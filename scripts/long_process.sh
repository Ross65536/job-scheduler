#!/bin/sh

steps="${1:-5}"
echo "sleeping for $steps steps"

i=1
until [ $i -gt $steps ]
do
  echo "Sleeping #$i"
  >&2 echo "test: Sleeping #$i"
  sleep 1
  ((i++))
done

echo "Done"