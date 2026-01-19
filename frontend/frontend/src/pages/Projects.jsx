import { useState, useEffect } from 'react';
import './pages.css';

function Projects() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Fetch projects from your Django API
    fetch('http://localhost:8000/api/repositories/') //Placeholder
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch projects');
        }
        return response.json();
      })
      .then(data => {
        setProjects(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className="page-container">
        <div className="page-content">
          <p className="loading">Loading projects...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page-container">
        <div className="page-content">
          <p className="error">Error: {error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-content">
        <h1 className="page-title">My Projects</h1>
        <p className="page-description">
          A collection of projects I've built, ranging from web applications to automation tools.
          Each project showcases different aspects of full-stack development and problem-solving.
        </p>

        <div className="projects-grid">
          {projects.map((project) => (
            <div key={project.id} className="project-card">
              <div className="project-header">
                <h2 className="project-name">{project.name}</h2>
                <span className="project-date">
                  {new Date(project.created_at).toLocaleDateString()}
                </span>
              </div>

              <p className="project-summary">
                {project.summary || 'No description available'}
              </p>

              {project.readme && (
                <div className="project-readme">
                  <details>
                    <summary>Read More</summary>
                    <div className="readme-content">
                      {project.readme}
                    </div>
                  </details>
                </div>
              )}

              <div className="project-tags">
                {/* Replace with actual tags when you add them */}
                <span className="tag">Python</span>
                <span className="tag">Django</span>
                <span className="tag">React</span>
              </div>

              <div className="project-footer">
                <span className="project-updated">
                  Updated: {new Date(project.updated_at).toLocaleDateString()}
                </span>
              </div>
            </div>
          ))}
        </div>

        {projects.length === 0 && (
          <p className="no-projects">No projects found. Check back soon!</p>
        )}
      </div>
    </div>
  );
}

export default Projects;