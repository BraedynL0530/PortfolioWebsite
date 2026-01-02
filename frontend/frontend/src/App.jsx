import { useState } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Terminal from './componets/terminal'
import Boring from './componets/boring-nav'
import AboutMe from './pages/AboutMe'; // add these later
import Projects from './pages/Projects';
import Contact from './pages/Contact';
import BTS from './pages/BTS';

function App() {
  const [isTerminal, setIsTerminal] = useState(true);
  return (

    <>
      <div className="header">
          <h1>Hi! im Braedyn!</h1>
          <p>I'm a full stack developer specialized in django and react</p>
          <button className="site-toggle" onClick={() => setIsTerminal(!isTerminal)}>{isTerminal ? "Boring Nav" : "Terminal nav"}</button>
          <div className="Social Buttons">
              <button>place holder</button>
          </div>
      </div>

      <Router>
          <Routes>
            <Route path="/" element={isTerminal ? (<Terminal/>) : (<Boring/> )} />
            <Route path="/about-me" element={<AboutMe />} />
            <Route path="/projects" element={<Projects />} />
            <Route path="/contact" element={<Contact />} />
            <Route path="/bts" element={<BTS />} />
            </Routes>
          </Router>


    </>
  )
}

export default App
