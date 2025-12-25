# Portfolio System – Architecture Overview

## Purpose

This portfolio represents the **pinnacle of my current technical ability** across:

- Frontend
- Backend
- Systems
- Machine Learning

It is **not a static website**, but an **interactive system** that demonstrates:

- Full-stack engineering
- Clear service boundaries
- Data pipelines
- Custom machine learning
- Intentional UI/UX design

The goal is to make my skills **immediately obvious** — no explanations, accounts, or setup required.

---

## High-Level Concept

The portfolio presents itself as a **desktop / terminal-inspired environment**  
(Aero / early Windows or riced Linux aesthetic).

Navigation is intentionally unconventional:

- Users can click UI elements (desktop-style)
- Or navigate via a terminal-like interface (`cd projects`, etc.)
- Subtle hints guide interaction without tutorials

This creates a **memorable, low-friction experience** while showcasing frontend skill.

---

## Frontend (React)

The frontend is responsible for:

- UI/UX
- Navigation and interaction
- Rendering project data
- Visual theming and animations

### Core Sections

- **About**
  - Name
  - Experience
  - Tech stack
- **Projects**
  - Interactive directory view
- **Contact**
  - Links and communication info

### Projects View

Repositories are displayed in a **grid layout**.

Each project is represented as a **card** containing:

- Title (top-left)
- Technology tags (near or above title)
- Short **ML-generated summary**
- Optional animated GIF background

Design details:

- Rounded corners
- Semi-transparent overlays for text readability
- Motion used intentionally, not excessively

The frontend performs **no heavy data processing** and relies entirely on APIs.

---

## Go Service – GitHub Ingestion Layer

A dedicated Go service handles all GitHub API interaction.

### Responsibilities

- Fetch repository metadata via  
  `https://api.github.com/users/<username>/repos`
- Retrieve README content when available
- Normalize and structure data
- Cache responses to avoid GitHub rate limits
- Expose a clean internal API for downstream services

### Why This Exists

- Isolates external API concerns
- Demonstrates concurrency and systems-level thinking
- Keeps the Django core focused and maintainable

---

## Django Core – Data & Orchestration

Django acts as the **central system coordinator**.

### Responsibilities

- Consume normalized repository data from the Go service
- Store repository metadata and derived features
- Coordinate ML inference
- Serve a stable API to the frontend

### Explicitly Not Responsible For

- Fetching GitHub data directly
- Heavy frontend logic

This separation enforces **clear service boundaries**.

---

## Machine Learning (PyTorch)

The system includes a **custom NLP pipeline** implemented in PyTorch.

### Goals

- Avoid reliance on large pretrained models
- Demonstrate understanding of ML fundamentals
- Maintain explainability and control

### Responsibilities

- Process README text
- Generate short project summaries
- Infer technology tags (e.g., Go, Django, React)

### Initial Approaches

- Tokenization and normalization
- TF-IDF or learned embeddings
- Lightweight neural models or classifiers
- Custom training loops and loss functions

Model outputs are **deterministic, debuggable, and explainable**.

---

## Data Flow Summary

1. Go service fetches and caches GitHub repository data
2. Django consumes structured repository data
3. README text is passed through the ML pipeline
4. Django stores enriched project data
5. React frontend renders the final interactive UI

---

## Design Principles

- No authentication required
- Zero friction for viewers
- Clear separation of concerns
- Explainable machine learning
- Shippable early, extensible later

---

## Non-Goals

- Large pretrained LLMs
- Overly complex DevOps or infrastructure
- User accounts or personalization
- Reinventing GitHub functionality
