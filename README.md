# Authentication Server

## Overview

The authentication system, handles access to the system.
Only social login is used to minimise any personally identifiable information, storing only the email address
which could potentially be used to identify a user. It also means we can dispose with the entire registration
process, email verification, password recovery and so on.

To maximise system security, we do not allow access to the internal system via websocket until we have already
got the user within the system.

## Authentication

We serve the authentication process over https with SSL and the websocket connections over wss with SSL
The client will use Google Sign In via the vue3-google-login module (or similar), on logging in, Google returns the
credentials to the client, the client then relays it to the authentication server via `/google/verify` url via a https
call in an Authorization header, we then verify the token data with Google Servers and create the user an uuid which 
acts as the users session id, if the email is new, a new user is added, if not, updated with the newly created session 
id, the session id is returned to the client.

The client then opens a websocket connection with the relay service, the service then checks the users ip address is not
in the IP ban list, if it is it closes the connection immediately rejecting the client before it can send any data to
the service, if the IP is not in the ban list, the service then sends an authentication verification request to the
authentication server/service and validates the session id passed to the service by the client, it returns a reply to
the relay service with true/false value having checked for banned status and banned the IP if the session is invalid
and not banned.

If the value is false, then the client connection closed, otherwise if valid we then replace the client connection in 
the connection pool with the validated session id as the new index for the client.

Within the system, to ensure correct ordering of messages and the delivery of service replies to the correct client, we
use the valid session id as the kafka message keys.

Any interruption to the websocket connection results in the connection closing and the client being deleted from the
client pool and requiring the client to login again to send and receive messages.

This forms the basis of a secure system.


## Considerations

The user is limited to one email address within the system, regardless of social login provider, for example the user
could have a social login with Google, then create a social login with Microsoft using their Google email address, that
is fine, and they could then log in with either Google or Microsoft, but logging in with a different email address would
result in a new account being created as the system would have no idea that the email given is related to the old email
used.

We do not suffer fools. The authentication process is an instant flow process, the session is passed on to the websocket 
on connection. The only way that the system could not validate the session correctly is because it is not a valid
session or the client is already banned. There is no way of spoofing it as unique time based random uuid's are used as
the session id's, hence on failure the ip address being banned, as it can only be a hack.

## Protection

There is no default route, only the route to verify.

The http server is constructed with slowris protection:

WriteTimeout: time.Second * 15,

ReadTimeout:  time.Second * 15,

IdleTimeout:  time.Second * 60,

The verification route is protected by token bucket rate limiting, 1 request per second per ip, the bucket clears by
the hour.

# Authentication Service

## Overview

So the authentication service is also the Authentication Service, it is a kafka producer and consumer, the service
currently services 2 commands:

- [x] authenticate 
- [x] ban_list

### authenticate

You heard about this above, authenticate receives the argument sessionId which carries the session id delivered to the
client from the `/google/verify` endpoint, it compares it to the database to validate the session id belongs to a user
and returns that status: true/false

### ban_list

This is an internal command used at relay service start up to obtain all the banned ip addresses to the connection can
be blocked if the client is already banned.
