import { useState } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Terminal from './componets/terminal'
import Boring from './componets/boring-nav'
//import AboutMe from './pages/AboutMe'; // add these later
//import Projects from './pages/Projects';
//import Contact from './pages/Contact';
//import BTS from './pages/BTS';

function App() {
  const [isTerminal, setIsTerminal] = useState(true);
  return (
    <>
      <Router>
        <div className="header">
          <div className="header-left">
            <h1>Hi! I'm Braedyn!</h1>
            <p>I'm a full stack developer specialized in Django and React</p>
          </div>

          <div className="header-right">
            <button className="site-toggle" onClick={() => setIsTerminal(!isTerminal)}>
              {isTerminal ? "Boring Nav" : "Terminal nav"}
            </button>

            <div className="social-buttons">
              <button>GitHub</button>
              <button>LinkedIn</button>
              <button>Twitter</button>
              <button>Email</button>
            </div>
          </div>
        </div>

        <Routes>
          <Route path="/" element={isTerminal ? <Terminal /> : <Boring />} />
        </Routes>
      </Router>
    </>
  )
}

export default App