#!/bin/bash

echo "grow-shrink"
echo "---"
NUM=$(( ( RANDOM % 10 )  + 1 ))
for (( c=1; c<=$NUM; c++ ))
do  
	echo "ohyeah up to $NUM"
done
