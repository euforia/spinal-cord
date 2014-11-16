===========
Spinal Cord
===========
Spinal cord is an event aggregator and reactor, which can perform arbitrary actions on a given event.

Installation
============
::

    go get github.com/euforia/spinal-cord/spinal-cord
    go get github.com/euforia/spinal-cord/nurv

This will generate 2 binaries that can be found under **$GOPATH/bin**


spinal-cord
===========

Server
------
The server is responsible for aggregating all input events via *PUSH/PULL* or *REQ/REP* and triggering/tasking workers to execute the appropriate handlers based on the fired event.  It also provides a basic web service and API to interact with the events and handlers.

Example::

    spinal-cord -handlers-dir /path/to/store/handlers/and/data \
                -webroot /absolute/path/to/webroot

Worker
------
The worker accepts tasks from the server and executes them.

Example::

    spinal-cord -worker \
                -handlers-dir /path/to/store/handlers/and/data


nurv
====

Example *PUSH/PULL*::

    nurv -type amqp -namespace openstack \
         -uri="amqp://guest:guest@rmq.open.stack.server.org:5672" \
         -queue="my.local.dev.info" \
         -routing-key="notifications.info" \
         -bind-to="keystone,nova,neutron"

In the above example *nurv* will connect to an openstack event queue and listen for events which in turn will be published to all the subscribers including the event handler system.


Example *REQ/REP*::

    nurv -type reqp \
         -namespace local \
         -event-type zmq.test \
         -data '{"client":"nurv"}'

In this example a single event will be sent to the system with the provided data.

**Note**: cross compilation is currently not supported due to CGO_ENABLED=1