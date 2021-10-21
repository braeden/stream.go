// When a client connects to a room (via whatever URL they tell us)

// We spawn a BLOCKING XREAD goroutine -> this routine will push the updates to the connected clients in the room

// When all clients exit the room, we kill the go routine, or when the stream dies we also kill the client

// scrolling should provide a different line reading functionality via API (basically the client should be able to ask for a range and get it back <1000)


// FE it should read the URL & fire off two requests, (lastest 100 and a subscription)


// Express server -> socket io library

// request to domain -> spawm webworker that connects to redis and XREADS in a loop
// have the worker send out socket events when it reads new data