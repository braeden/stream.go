const { createAdapter } = require("@socket.io/redis-adapter");
const { exit } = require("process");
const { createClient } = require("redis");
const app = require('express')();

const server = require('http').createServer(app);
const io = require('socket.io')(server, { cors: true });

const pubClient = createClient({ host: "localhost", port: 6379 });
const subClient = pubClient.duplicate();
const appClient = pubClient.duplicate();
const containsClient = pubClient.duplicate();


const port = process.env.PORT || 3002

appClient.on("error", err => {
    console.error(err);
    exit(1);
});

io.adapter(createAdapter(pubClient, subClient));

io.on('connection', socket => {
    console.log(`[${io.engine.clientsCount}] a user connected: ${socket.id}`);
    socket.on('disconnect', () => {
        // TODO: Redis unsubscribe when we're the last client to disconnect from a room
        console.log(`[${io.engine.clientsCount}] a user disconnected: ${socket.id}`);
    });

    socket.on('join', id => {
        console.log(id)
        containsClient.EXISTS(id, (err) => {
            console.log('exists', err)
            if (!err && id) {
                console.log('joined')
                socket.join(id);
                appClient.subscribe(`${id}-pubsub`)
            }
        })
    });

});

appClient.on('message', (channel, message) => {
    const id = channel.split('-')[0]
    // TODO: check that this checks size across adapters?
    if (!io.sockets.adapter.rooms.get(id)) {
        console.log('Last client disconnected')
        appClient.unsubscribe(channel);
    }
    io.to(id).emit('log', message)
    console.log(channel, message)
})

server.listen(port, () => console.log(`Socket server is up! ${port}!`));