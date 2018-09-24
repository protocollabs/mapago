# Info Request and Reply Server and Client

## Descriptions

- Start UDP and TCP listening on port (port) and listen for Info Request Messages
  - Wait for TCP Connection for IPv4 and IPv6
	- Wait for UDP Connection for IPv4 and IPv6
	- Wait for UDP Multicast Connection for IPv4 and IPv6
- Server answers with unicast message to the Sender using the original Protocol (UDP, TCP)
- Server assumes he support TCP THroughtput and UDP THroughput module
- CLient receives info reply and print it out on stdout.


## Tests

- Two Servers started on different PC, Client send request, two servers should be seen
