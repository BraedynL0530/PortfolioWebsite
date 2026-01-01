import { useState } from 'react'
import './App.css'
import Terminal from './componets/terminal'
import Boring from './componets/boring-nav'

function App() {
  const [isTerminal, setIsTerminal] = useState(true);
  return (
    <>
      <div className="header">
          <h1>Hi! im Braedyn!</h1>
          <p>I'm a full stack developer specialized in django and react</p>
          <button className="site-toggle" onClick={() => setIsTerminal(!isTerminal)}>{isTerminal ? "Boring Nav" : "Terminal nav"}</button>

        {isTerminal ? (<Terminal/>) : (<Boring/> )}
    </div>
    </>
  )
}

export default App
