Cleo
====

Cleo is an adventure into using Go for basic distributed systems tasks.
The goal of Cleo is to index and make large sets of League of Legends
gameplay information accessible for analysis. It currently utilizes Riot's
API to download a stream of game outcome data, builds an index of this
information, and runs a very simple query engine that can answer questions
of the form: 

	"How many games has <champion combination X> won against <champion combination Y>?"

It utilizes Google's protocol buffer library for some serialization tasks
and uses MongoDB as a backend for storing extracted game data.

Basic architecture
====================
FETCHER: retrieving the data
Fetcher is a daemon that regularly accesses Riot's game history API to
retrieve information for known summoners. It starts off with a seed set
of summoners and updates its list as it discovers new summoners in new
games.

When fetcher receives game information it writes it parses it into an
internal data format and stores the parsed version in a MongoDB instance.

PACKER: building the index
Packer reads the MongoDB instance and reformats + filters the games into
an effecient data structure that Lolstat can use for searching.

LOLSTAT: searching
Lolstat reads in the index generated by the packer and opens up a network
connection to await queries. When queries come in it performs searches
on the index, computes key statistics, and returns them in a query response.

Usage
=========
Note that cleo is currently specifically designed for League of Legends, though it
shouldn't be conceptually difficult to transition each component to work for other
data sources. I want to make this more true over time. The following instructions
are for using cleo specifically for its currently intended purpose stated above.

1) Install all prerequisites:
  * golang
  * Protocol buffer compiler and [goprotobuf](https://code.google.com/p/goprotobuf/)
  * MongoDB server and [mgo](http://labix.org/mgo) client driver

2) Whitelist a League of Legends account for access to the [Riot API](http://developer.riotgames.com/)

3) Compile protocol buffers by running ./build from the base directory. You may need to customize the
paths inside of the script depending on where you installed goprotobuf in step 1.

3) Build all of the executables.
  * go build fetcher
  * go build packer
  * go build lolstat
  * go build frontend

4) Run fetcher to build up a list of games. Fetcher will require access to a MongoDB instance.

	./fetcher -apikey=<your_api_key>

5) Once you have an adequate number of games available, run:

	./packer -apikey=<your_api_key>

This will generate a packed "pcgl" file that can be used for serving. Note that it will also generate some
static resources based on information from the Riot API and game list that will be consumed by the frontend.
Whenever you'd like to rebuild a new pcgl file, just rerun packer. It should also be rerun whenever Riot
adds new champions to add support for them in the frontend.

6) The serving system can be started with:

	./lolstat (to start the backend service)
	./frontend (to start the frontend web server that talks to lolstat)

7) You can view the frontend by visiting http://<domain>:8088/ in your favorite (Angular-supported) web browser.
