import React, { useState, useEffect } from 'react';
import socketIOClient from "socket.io-client";
import { useLocation } from 'react-router-dom'

const ENDPOINT = "http://127.0.0.1:3002";


const Log = () => {
    const [response, setResponse] = useState<string[]>([]);
    const { pathname } = useLocation();

    useEffect(() => {
        const socket = socketIOClient(ENDPOINT);
        socket.emit('join', pathname.slice(1))
        socket.on('log', data => {
            setResponse((old) => [...old, data]);
        });
    }, [pathname]);
    return (
        <div>Hello World {response.join('\n')}</div>
    );

}
export default Log;
