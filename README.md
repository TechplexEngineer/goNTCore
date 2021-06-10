# goNTCore

GoLang implementation of the WPILib ntcore software.

## What is Network Tables?
Network Tables is a network protocol for a key-value store which can be read from and 
written to by multiple remote clients. A central server, most often running on a 
FIRST FRC robot controller, is responsible for providing information consistency 
and for facilitating communication between clients.

## Network Table V3

To view the standards that this package follows read [NetworkTables v3 Spec](https://github.com/wpilibsuite/allwpilib/tree/main/ntcore/doc)

## TODO (possibly out of date)

- [x] Create all the possible Entry types
  - Include a way to Marshal and Unmarshal them
- [x] Create all the possible Message types
  - Include a way to Marshal and Unmarshal them
- [x] Create a NetworkTable interface that should be used for anything to talk on the Network
- [x] Create a Data interface for NetworkTable interface to use for manipulating data
- [x] Create a NetworkTable struct that implements the NetworkTable interface but consumes the Data interface
- [x] Create a Server that can listen to multiple clients as well as broadcast to them
  - Cleanup old connections
- [x] Create a Client that can listen to a server for messages
  - Should be able to initiate handshake
- [ ] Create a data struct to implament Data that is connected to a Key Value store
  - Has to handle keys, IDs and SNs correctly
  - Should handle persistent flag correctly
- [ ] Server needs to have NetworkTabler methods so that it can send the appropriate messages
- [ ] Client needs to have NetworkTabler methods so that it can send the appropriate messages
- [ ] Fix the Server handshake to return EntryAssign for all entries it knows about
- [ ] Fix the Client handshake to return what is different between what the server has and what it has
- [ ] Modify Server handler method to manipulate data on the server side
- [ ] Modify Client handler method to manipulate data on the client side
