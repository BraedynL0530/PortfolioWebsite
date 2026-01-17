// Boring.jsx
import { useNavigate } from 'react-router-dom';
import './boring.css'

function Boring() {
  const navigate = useNavigate();

  function handleNavClick(page) {
    navigate(`/${page}`);
  }

  return (
    <>
      <div className="boring-container">
        <header className="boring-header">
          <h1 className="boring-title">Braedyn</h1>
          <p className="boring-subtitle">15 year old fullstack developer based in NC</p>
        </header>

        <nav className="nav">
          <button onClick={() => handleNavClick('about-me')} className="nav-btn">
            <span className="nav-icon">👤</span>
            <span className="nav-text">About me</span>
            <span className="nav-desc">Learn about my background</span>
          </button>
          <button onClick={() => handleNavClick('projects')} className="nav-btn">
            <span className="nav-icon">💻</span>
            <span className="nav-text">Projects</span>
            <span className="nav-desc">Check out what I've built</span>
          </button>
          <button onClick={() => handleNavClick('contact')} className="nav-btn">
            <span className="nav-icon">📧</span>
            <span className="nav-text">Contact</span>
            <span className="nav-desc">Get in touch with me</span>
          </button>
          <button onClick={() => handleNavClick('bts')} className="nav-btn">
            <span className="nav-icon">🔧</span>
            <span className="nav-text">BTS</span>
            <span className="nav-desc">How this site works</span>
          </button>
        </nav>
      </div>
    </>
  )
}

export default Boring