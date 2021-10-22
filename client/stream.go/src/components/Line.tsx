
interface LineInterface {
    text: string;
    line: number;
    id: string;
}

const Line = ({text, line, id}: LineInterface) => {
    return (
        <tr>
            <td> aria-label={line}</td>
            <td><pre><code>{text}</code> </pre></td>
            
        </tr>
    
    );

}
export default Line;
