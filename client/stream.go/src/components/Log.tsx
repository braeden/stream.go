import React, { useState, useEffect } from 'react';
import socketIOClient from "socket.io-client";
import { useLocation } from 'react-router-dom'

const ENDPOINT = "http://127.0.0.1:3002";
interface Line {
    text: string;
    line: number;
    id: string;
}

const Log = () => {
    const [response, setResponse] = useState<Line[]>([]);
    const [socket, setSocket] = useState<any>();
    const [id, setId] = useState<string>();


    const { pathname } = useLocation();

    useEffect(() => {
        setSocket(() => socketIOClient(ENDPOINT));
        setId(() => pathname.slice(1));
    }, []);

    const getPrev = (count = 100) => {
        socket.emit('askRange', JSON.stringify({
            id, count, earliestStreamId: response[0]?.id
        }))
    }

    useEffect(() => {
        if (socket && id) {
            socket.emit('join', id)

            socket.on('respondRange', (data: string) => {
                if (JSON.parse(data).err) {
                    return console.error(data)
                }
                console.log(data)
                const lines: Line[] = JSON.parse(data)
                console.log(lines)
                setResponse((old) => [...lines, ...old]);
            });

            socket.on('log', (data: string) => {
                const line: Line = JSON.parse(data)
                setResponse((old) => [...old, line]);
            });
        }
    }, [socket, id])



    return (
        <div>Hello World <button onClick={() => getPrev()}>prev</button>{response.map(e => JSON.stringify(e)).join('\n')}</div>
    );

}
export default Log;
