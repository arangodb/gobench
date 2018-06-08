# Test results for current devel

For this test we run on c12 and use a server with exactly 10 threads in
the scheduler. We vary across the following variables:

  - RocksDB/MMFiles
  - HTTP/VST
  - TCP/TLS
  - parallelism 1, 2, 4, 8, 16 and 32

For HTTP, we allow 25 parallel connections, for VST we allow 3 parallel
connections.

## MMFiles/HTTP/TCP

25 connections

    Parallelism Total           Median          99.9%
    1           43.811958866s   425.539us       1.468269ms
    2           23.693077593s   455.973us       1.61635ms
    4           13.031166124s   491.806us       1.839028ms
    8           6.430560719s    421.225us       1.989469ms
    16          11.022556414s   1.762723ms      4.077076ms
    32          13.691797844s   4.29705ms       9.554592ms

## MMFiles/HTTP/TLS

25 connections

    Parallelism Total           Median          99.9%
    1           47.233522553s   456.168us       1.525322ms
    2           25.597244453s   490.431us       1.769503ms
    4           13.833053105s   520.475us       2.265723ms
    8           7.566009259s    437.313us       2.690544ms
    16          9.306772712s    1.272314ms      7.971422ms
    32          9.971594544s    2.876367ms      11.405024ms

## MMFiles/VST/TCP

3 connections

    Parallelism Total           Median          99.9%
    1           38.844317146s   380.503us       1.181731ms
    2           26.702249139s   425.235us       42.511094ms
    4           16.121206503s   410.219us       43.715977ms
    8           8.602968745s    358.148us       43.716838ms
    16          3.996344563s    304.775us       43.848707ms
    32          2.923426685s    392.298us       44.25031ms

## MMFiles/VST/TLS

3 connections

    Parallelism Total           Median          99.9%
    1           47.726644061s   467.244us       1.224863ms
    2           30.064477403s   502.857us       42.577395ms
    4           18.180501584s   466.019µs      43.878219ms
    8           8.894551513s    407.267us       43.767553ms
    16          4.450981859s    369.668us       44.044561ms
    32          4.007825122s    495.039us       44.88047ms

## RocksDB/HTTP/TCP

25 connections

    Parallelism Total           Median          99.9%
    1           52.14709011s    507.188us       1.528656ms
    2           29.17723858s    562.583us       1.70891ms
    4           15.718913501s   604.841us       1.892398ms
    8           7.534067742s    526.856us       2.012215ms
    16          10.94720117s    1.758492ms      4.106403ms
    32          13.696427053s   4.280116ms      9.432008ms

## RocksDB/HTTP/TLS

25 connections

    Parallelism Total           Median          99.9%
    1           56.316947617s   548.176us       1.582011
    2           31.234825537s   603.712us       1.835642ms
    4           16.362389804s   626.639us       2.225103ms
    8           8.743173047s    535.154us       2.681041ms
    16          8.087401563s    1.136756ms      5.809873ms
    32          9.855540795s    2.819578ms      12.652778ms

## RocksDB/VST/TCP

3 connections

    Parallelism Total           Median          99.9%
    1           48.407452799s   477.137us       1.128316ms
    2           28.91174666s    533.952us       1.737154ms
    4           17.711946594s   495.465us       43.710137ms
    8           9.015320692s    455.806us       43.712926ms
    16          4.238956393s    386.517us       43.655509ms
    32          3.519079961s    548.992us       44.186697ms

## RocksDB/VST/TLS

3 connections

    Parallelism Total           Median          99.9%
    1           59.709246841s   584.649us       1.262538ms
    2           35.243689829s   626.462us       42.133326ms
    4           21.145276742s   621.615us       43.915448ms
    8           9.59612343s     521.523us       43.698952ms
    16          5.099449695s    449.953us       44.017985ms
    32          4.351975771s    757.976us       45.240094ms

