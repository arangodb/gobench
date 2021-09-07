# Testplan

Use postDocs, and seedDocs and readDocs, so two types of measured requests.

Use single server.

Use TLS throughout.

Do all measurements with HTTP/1+VPACK, HTTP/2+VPACK, VST+VPACK.

Check the following combinations of

parallelism      number of connections        
1                1
2                2
4                4
8                8
16               1
16               4
16               16
128              128

Use 100000 docs or so to speed up tests.

Use arangosh, use go-driver.
