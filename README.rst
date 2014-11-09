===========
Spinal Cord
===========
Spinal cord is an event aggregator and reactor, which can perform arbitrary actions on a given event.


spinal-cord
===========

Server
------
The server is responsible for aggregating all input events via PUSH/PULL or REQ/REP and triggering/tasking workers to execute the appropriate handlers based on the fired event.

Example::

    spinal-cord -handlers-dir /path/to/store/handlers/and/data -webroot /absolute/path/to/webroot

Worker
------
The worker accepts tasks from the server and executes them.

Example::

    spinal-cord -worker -handlers-dir /path/to/store/handlers/and/data


nurv
====

Example 1::

    nurv -type amqp -namespace openstack -uri="amqp://guest:guestpass@app.stack.dev1.cloudsys.tmcs:5672" -queue="my.local.dev.info" -routing-key="notifications.info" -bind-to="keystone,nova,neutron"

Example 2::

    nurv -type reqp -event-type zmq.test -namespace local -data '{"client":"nurv"}'