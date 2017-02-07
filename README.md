A Tale of Two Queues
====================

Related source code from a blog post
[A Tale of Two Queues](http://blog.jupo.org/2013/02/23/a-tale-of-two-queues/),
benchmarking [Redis](http://redis.io) and [ZeroMQ](http://www.zeromq.org)
pub-sub, using [Python](http://python.org) and Google's
[Go language](http://golang.org).

Software and Libraries
======================

Following are the software and library requirements for running these
benchmarks, using [homebrew](http://mxcl.github.com/homebrew) for OSX,
[pip](http://www.pip-installer.org) for Python and [go get](
http://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies)
for Go. Installation should be similar on various Linuxes using their respective
package managers rather than homebrew. You'll also need to run Redis and Mosquitto once
installed.

    $ brew install zeromq [1]
    $ brew install redis [2]
    $ brew install mosquitto [3]
    $ brew install gnuplot [4]
    $ pip install pyzmq [5]
    $ pip install redis [6]
    $ pip install hiredis [7]
    $ pip install paho-mqtt [8]
    $ go get github.com/garyburd/redigo/redis [9]
    $ go get github.com/pebbe/zmq4 [10]
    $ go get -u github.com/go-mangos/mangos [11]

1. <http://www.zeromq.org>
2. <http://redis.io>
3. <https://mosquitto.org>
4. <http://www.gnuplot.info>
5. <https://github.com/zeromq/pyzmq>
6. <https://github.com/andymccurdy/redis-py>
7. <https://github.com/pietern/hiredis-py>
8. <https://pypi.python.org/pypi/paho-mqtt>
9. <https://github.com/garyburd/redigo>
10. <https://github.com/pebbe/zmq4>
11. <https://github.com/go-mangos/mangos>

Running
=======

To run the test use `python bench.py`. Before running the benchmark start the brokers:

* `redis-server`
* `go run run_broker.go`
* `go run run_broker.go --mangos`

Currently the `bench.py` script does not start the brokers for `zmq` and `mangos` automatically.

Troubleshooting
===============

Increase the number op open files in OSX using: `ulimit -n 4096`

Props
=====

Based upon the original from [Stephen](https://github.com/stephenmcd/two-queues)

Stuff used from work of others
  * [Redis 2.6.10 vs ZeroMQ 3.2.3](https://github.com/stephenmcd/two-queues/issues/1)
  * [MQTT broker addition](https://github.com/stephenmcd/two-queues/pull/5)

