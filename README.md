# Authentication Server

## Overview

The authentication system, handles access to the system.
Only social login is used to minimise any personally identifiable information, storing only the email address
which could potentially be used to identify a user. It also means we can dispose with the entire registration
process, email verification, password recovery and so on.

To maximise system security, we do not allow access to the internal system via websocket until we have already
got the user within the system.

For this reason, we serve the authentication process over standard https.
When the client is logged in, the token data will be returned to the client which it then passed onto the websocket
server, we will (at point of providing the redirect url for the social login) have stored the users ip address against
a timeout, the client then has 10 minutes to access the system to confirm the details.
The client passes the token data to the service, we then verify those details with the social login provider, if they
are valid, access is permitted to the system, if not the client is booted immediately and the ip permanently banned.

When the token details are verified, the reply from the service includes the users' id, this is then stored against the
client and used for services to identify the user when linkage is required, the client does not however see the user id
ever.

The user is limited to one email address within the system, regardless of social login provider, for example the user
could have a social login with Google, then create a social login with Microsoft using their Google email address, that
is fine, and they could then log in with either Google or Microsoft, but logging in with a different email address would
result in a new account being created as the system would have no idea that the email given is related to the old email
used.


