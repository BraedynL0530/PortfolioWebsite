import { useState, useEffect } from 'react';
import './pages.css';

function AboutMe() {
  return (
    <div className="page-container">
      <div className="page-content">
        <h1 className="page-title">About Me</h1>

        <section className="about-section">
          <div className="about-text">
            <h2>Who I Am</h2>
            <p>
              Hey! I'm Braedyn, a 15-year-old full-stack developer based in North Carolina.
              I specialize in building  web applications using Django and React.
            </p>

            <h2>My Journey</h2>
            <p>
              [Your description here - talk about how you got into programming, what drives you,
              your goals, etc.]
            </p>

            <h2>Skills & Technologies</h2>
            <div className="skills-grid">
              <div className="skill-category">
                <h3>Frontend</h3>
                <ul>
                  <li>React</li>
                  <li>JavaScript/TypeScript</li>
                  <li>HTML/CSS</li>
                </ul>
              </div>

              <div className="skill-category">
                <h3>Backend</h3>
                <ul>
                  <li>Django</li>
                  <li>Python</li>
                  <li>REST APIs</li>
                  <li>Golang</li>
                </ul>
              </div>

              <div className="skill-category">
                <h3>Tools</h3>
                <ul>
                  <li>Git/GitHub</li>
                  <li>VS Code</li>
                  <li>Linux</li>
                  <li>Docker</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="stats-container">
            <h2>GitHub Stats</h2>
            <div className="stat-placeholder">
              <p>[GitHub contribution stats will go here]</p>
              <img src="https://via.placeholder.com/400x200/1a1a1a/00ff88?text=GitHub+Stats" alt="GitHub Stats Placeholder" />
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}

export default AboutMe;