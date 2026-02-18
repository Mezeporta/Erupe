// Package entranceserver implements the MHF entrance server, which listens on
// TCP port 53310 and acts as the gateway between authentication (sign server)
// and gameplay (channel servers). It presents the server list to authenticated
// clients, handles character selection, and directs players to the appropriate
// channel server.
//
// The entrance server uses MHF's custom "binary8" encryption and "sum32"
// checksum for all client-server communication. Each client connection is
// short-lived: the server sends a single response containing the server list
// (SV2/SVR) and optionally user session data (USR), then closes the
// connection.
package entranceserver
