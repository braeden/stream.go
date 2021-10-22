import React, { useState, useEffect } from 'react';
import socketIOClient from "socket.io-client";
import { useLocation } from 'react-router-dom'
import Line from './Line'
import { XTerm } from 'xterm-for-react';

const ENDPOINT = "http://127.0.0.1:3002";
interface LineInterface {
    text: string;
    line: number;
    id: string;
}

const Log = () => {
    const [response, setResponse] = useState<LineInterface[]>([]);
    const [socket, setSocket] = useState<any>();
    const [id, setId] = useState<string>();
    const xtermRef = React.useRef<any>(null);





    const { pathname } = useLocation();

    useEffect(() => {
        setSocket(() => socketIOClient(ENDPOINT));
        setId(() => pathname.slice(1));
    }, []);

    const writeLine = (line: LineInterface) => {
        if (id) {
            const date = new Date(Number(line.id.split('-')[0]))
            xtermRef?.current?.terminal?.writeln(`${date.toISOString()} ${line.text}`)
        }
    }

    const getPrev = (count?: number) => {
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

                const lines: LineInterface[] = JSON.parse(data)
                lines.forEach(e => writeLine(e))
                // setResponse((old) => [...lines, ...old]);
            });

            socket.on('log', (data: string) => {
                writeLine(JSON.parse(data))
                // setResponse((old) => [...old, line]);
            });

            getPrev()
        }
    }, [socket, id])

    return (
        <div>
            <XTerm ref={xtermRef} />
        </div>
    );

}
export default Log;
