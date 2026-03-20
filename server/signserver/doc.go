// Package signserver implements the MHF sign server, which handles client
// authentication, session creation, and character management. It listens
// on TCP port 53312 and is the first server a client connects to in the
// three-server network model (sign, entrance, channel).
package signserver
