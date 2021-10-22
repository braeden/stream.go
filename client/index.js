const { createAdapter } = require("@socket.io/redis-adapter");
const { exit } = require("process");
const { createClient } = require("redis");
const app = require('express')();

const server = require('http').createServer(app);
const io = require('socket.io')(server, { cors: true });

const pubClient = createClient({ host: "localhost", port: 6379 });
const subClient = pubClient.duplicate();

const appSubClient = pubClient.duplicate();
const appClient = pubClient.duplicate();

const port = process.env.PORT || 3002

subClient.on('error', err => {
    console.error(err);
    exit(1);
});

(async () => {
    await appClient.connect();
    await pubClient.connect();
    await subClient.connect();
    await appSubClient.connect();
})();
const handler = (message, channel) => {
    const id = channel.split('-')[0]
    // TODO: check that this checks size across adapters?
    // if (!io.sockets.adapter.rooms.get(id)) {
    //     console.log('Last client disconnected')
    //     await subClient.unsubscribe(channel);
    // }
    io.to(id).emit('log', message)
};
// Renable this when io adapter updates
// io.adapter(createAdapter(pubClient, subClient));

io.on('connection', socket => {
    console.log(`[${io.engine.clientsCount}] a user connected: ${socket.id}`);
    socket.on('disconnect', () => {
        // TODO: Redis unsubscribe when we're the last client to disconnect from a room
        console.log(`[${io.engine.clientsCount}] a user disconnected: ${socket.id}`);
    });

    socket.on('join', async id => {
        if (await appClient.EXISTS(id)) {
            console.log('joined')
            socket.join(id);
            await appSubClient.subscribe(`${id}-pubsub`, handler)
        }
    });

    socket.on('askRange', async data => {
        if (data) {
            try {
                const { id, earliestStreamId = '+', count: COUNT } = JSON.parse(data)
                // TODO: Validate that socket is in the room it requests
                // const socketInRoom = await io.of('/').adapter.sockets(new Set([id])).has(socket.id)
                if (!id || !(await appClient.EXISTS(id))) {
                    throw ('Invalid room or arguments');
                }
                const res = COUNT ?
                    await appClient.XREVRANGE(`${id}-stream`, earliestStreamId, '-', { COUNT })
                    : await appClient.XREVRANGE(`${id}-stream`, '+', '-')
                const sorted = res.map(({ id, message }) => ({ ...message, id })).sort((a, b) => a.line - b.line)
                // We may want to sort by redis id (unix timestamp - #) instead of line
                if (earliestStreamId !== '+') {
                    sorted.pop() // Non-inclusive if we don't get asked from the start
                }
                socket.emit('respondRange', JSON.stringify(sorted))
            } catch (e) {
                console.error(e)
                socket.emit('respondRange', JSON.stringify({ err: e }))
            }
        }
    });

});

// appSubClient.on('message', (channel, message) => {
//     console.log('here')
//     const id = channel.split('-')[0]
//     // TODO: check that this checks size across adapters?
//     if (!io.sockets.adapter.rooms.get(id)) {
//         console.log('Last client disconnected')
//         subClient.unsubscribe(channel);
//     }
//     io.to(id).emit('log', message)
// })

server.listen(port, () => console.log(`Socket server is up! ${port}!`));