import './terminal.css'
import {useState} from "react";


function Terminal() {


    const validCommands = {
        'cd': 'Change page',
        'ls': 'List available pages',
        'clear': 'Clear terminal',
        'help': 'Show available commands',
    };
    const validPages = {
        'about me': 'A brief description of myself and my background',
        'projects': 'My current projects and a brief summary of each',
        'contact': 'Contact info, email, linkdin, etc',
        'bts': 'Behind The Scenes, How this website works!'
    }
    const [history, setHistory] = useState([])
    const [input, setInput] = useState('');
    const currentPath = "C:/Users/Braedyn/Desktop/Portfolio"
    const asciiCat = `
  ／l、
 (ﾟ､ ｡ ７
  l  ~ヽ
  じしf_,)ノ
`;

    function getCommandColor(text) {
        const parts = text.trim().split(' ');
        const command = parts[0].toLowerCase();
        const arg = parts[1]?.toLowerCase();

        if (validCommands[command]) {
            // Check if cd has valid arg
            if (command === 'cd' && arg) {
                return Object.keys(validPages).includes(arg) ? 'valid' : 'error';
            }
            return 'valid';
        }

        if (command === '') return 'default';
        return 'invalid';
    }

    function handleCommand(input) {
        const parts = input.trim().split(' ');
        const command = parts[0].toLowerCase();
        const arg = parts[1];

        let output = { command: input, type: 'default', result: '' };

        switch (command) {
            case 'help':
                output.type = 'info';
                output.result = Object.entries(validCommands)
                    .map(([cmd, desc]) => `  ${cmd.padEnd(10)} - ${desc}`)
                    .join('\n');
                break;

            case 'cd':
                if (!arg) {
                    output.type = 'error';
                    output.result = 'cd requires a page name';
                } else if (Object.keys(validPages).includes(arg.toLowerCase())) {
                    output.type = 'info';
                    output.result = `Opening ${arg}...`;
                    // redirect after a small delay for UX
                    setTimeout(() => {
                        window.location.href = `/${arg.replace(/ /g, '-')}`;
                    }, 300);
                } else {
                    output.type = 'error';
                    output.result = `No such page: ${arg}`;
                }
                break;

            case 'ls':
                output.type = 'info';
                output.result = Object.entries(validPages)
                    .map(([page, desc]) => `${page.padEnd(15)} - ${desc}`)
                    .join('\n');
                break;
            case 'clear':
                setHistory([]);
                break;
            default:
                output.type = 'error';
                output.result = `Command not found: ${command}. Type 'help' for available commands.`;
        }
        setHistory([...history, output]);
        setInput('');
    }

    function handleKeyDown(e) {
        if (e.key === 'Enter') {
            handleCommand(input);
        }
    }
    return (
    <>
        <div className="terminal">
            <div className="terminal-header">
                <span className="circle red"></span>
                <span className="circle yellow"></span>
                <span className="circle green"></span>
            </div>

            <div className="terminal-body">
                <pre className="ascii-cat">{asciiCat}</pre>                    {history.map((entry, i) => (
                        <div key={i} className="history-entry">
                            <div className="history-command">
                                <span className="prompt">{currentPath}$ </span>
                                <span className={`command-text ${entry.type}`}>
                                    {entry.command}
                                </span>
                            </div>
                            {entry.result && (
                                <pre className={`output ${entry.type}`}>
                                    {entry.result}
                                </pre>
                            )}
                        </div>
                    ))}
                </div>

            <div className="terminal-input">
                <span className="prompt">{currentPath}$ </span>
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={handleKeyDown}
                        className={getCommandColor(input)}
                        autoFocus
                    />
                </div>
        </div>
    </>
    )
}
export default Terminal