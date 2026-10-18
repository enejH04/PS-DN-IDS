#!/bin/bash

N=$1

for ((i=0; i<N; i++)) {
    ./naloga-6 -id "$i" -n "$N" -rid $(( N - 1 )) &
}

wait