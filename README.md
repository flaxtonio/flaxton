# Flaxton: Low level automated load balancer
<p>
Flaxton is an open source project to automatically load balance Docker containers using Linux ipTables, for getting 
maximum performance and scalability.
</p>
<p>
Using Flaxton you need just to scale the number of your containers, and you don't need to reload any configuration file
for Load Balancing as you are doing now with other TCP load balancers by adding new container's IP addresses.
</p>
<p>
Flaxton consists of 3 parts:
</p>
<ol>
  <li><b>Flaxton Daemon:</b> background working process which will automatically detect and load balance Docker containers (images)
  and child servers. Flaxton Daemon making Load Balancing by adding DNAT rules to ipTables.
  </li>
  <li><b>Flaxton CLI:</b> small program for making manipulations with Flaxton Daemons, without dealing with server IP addresses
  </li>
  <li><b>Flaxton Daemon Backend:</b> Node.js based (probably will switch to Golang) server side program to handle Flaxton Daemon
  requests with Docker container monitoring information, and task management for Flaxton Daemon.  Main functionality is to store 
  Daemon information to MongoDB and send available tasks from database to Flaxton Daemon.
  </li>
</ol>

# Installation
<p>
Flaxton Daemon + CLI is a single executable which you can find in <code>bin</code> directory, it compiled for x64 systems MAC OS and Linux.
</p>
<p>
If you want to compile it manually you need to have installed and configured <a href="http://golang.org" target="_blank">Go programming language</a>.
<br/>
You can find <a href="http://golang.org" target="_blank">Go programming language</a> installation instructions at <a href="https://golang.org/doc/install" target="_blank">https://golang.org/doc/install</a>
</p>
<p>
After installing and having configured <code>golang</code> you can compile Flaxton project using following simple steps:
</p>
```bash
git clone https://github.com/flaxtonio/flaxton
cd flaxton
export GOPATH=$(pwd)
# installing dependencies
go get
cd src
# after this command you will have flaxton executable in src directory
go build flaxton.go
```

# Flaxton Daemon
<p>
Before starting Flaxton Daemon you need to login using this command <br/>
<code>./flaxton login -u test -p test</code> it will create <code>.flaxton</code> configuration json file in your <code>/home/[user]</code>
directory, which is need to make a secure query's to <code>container.flaxton.io</code> repository.
</p>
<p>
To start Flaxton Daemon you need to type one of the following commands
</p>
```bash
sudo ./flaxton daemon
sudo ./flaxton -d
```
<p>
It will start tracking available Docker images and containers with their state/load information.<br/>
And on every 1 second Flaxton Daemon will ask container.flaxton.io server to send him available tasks, by sending 
tracked information from Docker (containers with stat/load, images).
</p>
<p>
Flaxton Daemon will start Load Balancing for specific port only when he will recieve Task, with 3 parameters
</p>
<ol>
  <li>Port for load balancing</li>
  <li>Docker image name for load balancing containers created only from that image<br/>
    Or Child server Ip address who could also be load balancing with given port
  </li>
  <li>Container Port - Forward requests to port on Docker container
    <br/>
    Or Port on child server
  </li>
</ol>

# Flaxton CLI
<p>
Flaxton CLI is used for sending Tasks to container.flaxton.io repository, which will be available for executing on specific Flaxton Daemon.
</p>
<p>
Here is the basic usages for Flaxton CLI
</p>
```bash
# Getting list of available Flaxton Daemon servers for logged in user
./flaxton daemon list

# Setting name for Daemon server
./flaxton daemon set_name <daemon_id> <daemon_name>

# Transfering Docker image to one of the Daemon servers. It will transfer flaxton/php:test image to
# daemon_dev server and will start 2 containers with command "/run.sh"
./flaxton transfer -daemon daemon_dev -img flaxton/php:test -cmd /run.sh -count 2

# Starting Load Balancer for Docker image, from Host OS 80 port to Containers 80 port
# This will start load balancing only for containers created from flaxton/php:test image
./flaxton daemon map-image daemon_dev 80 flaxton/php:test 80


# Stoping load balancer for specific port
./flaxton daemon stop-port daemon_dev 80
```

<p>
<b>Development process very dynamic and many commands is adding day by day !</b>
</p>

# Flaxton Daemon Backend
<p>
This part is not open sources yet, because it will have a lot of changes during this month.<br/>
At this moment it is writtent in Node.js and it's live on container.flaxton.io with MongoDB database.
</p>
<p>
We are thinking of switching Node.js backend to Go project.
</p>

# Questions ?
<p>
This technology is giving a lot of opportunity and freedom for building Infinity scalable Docker infrastructure,
because you don't need to deal with Docker containers networking, all that is doing Flaxton Daemon by manipulating Linux ipTables.
</p>
<p>
<strong><i>One of the biggest advantages of Flaxton Technology is that it's not passing traffic through it as it
is doing other TCP load balancers, it is just manipulating Linux iptables. And for Flaxton Daemon it doesn't matter you have 1K req/s or 800k req/s,
 it wouldn't affect CPU or RAM usage of Flaxton Daemon.</i></strong>
</p>

<p>
If you have a questions lets talk about <a href="mailto:tigran@flaxton.io">tigran@flaxton.io</a>
</p>
