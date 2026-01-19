import './pages.css';

function BTS() {
  return (
    <div className="page-container">
      <div className="page-content">
        <h1 className="page-title">Behind The Scenes</h1>

        <section className="bts-section">
          <h2>How This Site Works</h2>
          <p>[Your description of the tech stack and architecture]</p>

          <h2>Tech Stack</h2>
          <p>[Explain your choices - why React, why Django, etc.]</p>

          <h2>Interesting Challenges</h2>
          <p>[Talk about problems you solved while building this]</p>
        </section>
      </div>
    </div>
  );
}

export default BTS;