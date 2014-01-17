/*Package minigrush is a simplified version of the HTTP relayer https://github.com/telefonicaid/Rush

There are a few entities cooperating for relaying HTTP requests to the target host specified
in the header field 'X-Relayer-Host' from the incoming request. Listener and Consumer do the main job
accepting request, enqueueing them in a shared channel and storing the response from the target host
(or any errors that may occur). Replyer gives back those responses to the client. Recoverer takes incoming request
Recoverer gets incoming requests that are stored but failed to make and enqueues them again

Listener accepts incoming HTTP requests and stores them in a queue, after some processing
a request become a Petition. It assigns an ID to the petition and returns the ID to the client

Consumer takes a petition from the queue and makes the meant request to the target host.
The response of the target host is stored as a Reply so that it can later be accessed by the client,
using the ID generated in the request to the listener


*/
package minigrush
