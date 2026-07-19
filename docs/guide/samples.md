# Sample Projects

VisionStudio includes sample projects to help you learn the workflow and see best practices in action.

## Overview

Sample projects provide:

- Complete example configurations
- Pre-built V2MOMs, capabilities, and roadmaps
- Reference maturity models
- Templates you can adapt for your projects

## Available Samples

### Grafana Sample

A comprehensive example based on a platform engineering scenario.

| Component | Contents |
|-----------|----------|
| V2MOM | Organization + 5 team V2MOMs |
| Capabilities | 5 domains with multiple capabilities |
| Roadmap | Multi-quarter initiative planning |
| Maturity Model | 5-dimension assessment |

**Domains:**

- Core Platform
- API/DX
- Operations
- Quality
- Security

### Simple Sample

A minimal example for learning the basics.

| Component | Contents |
|-----------|----------|
| V2MOM | Single project V2MOM |
| Capabilities | Single domain |
| Roadmap | Simple initiative |
| Maturity Model | Basic 3-dimension model |

## Using Sample Projects

### Browsing Samples

1. Click **Samples** in the sidebar (or use Add Project)
2. Browse available sample projects
3. View sample details and file counts
4. Preview README for each sample

### Importing a Sample

1. Select a sample project
2. Click **Import Sample**
3. Choose destination directory
4. Sample files are copied to your workspace
5. Project appears in your project list

### Learning from Samples

After importing:

1. Explore the V2MOM structure
2. Review capability organization
3. See how roadmap links to V2MOM
4. Examine maturity model configuration
5. Use as templates for your own projects

## Sample Structure

Each sample follows standard project structure:

```
sample-name/
├── project.json           # Project metadata
├── README.md              # Sample documentation
├── v2mom/
│   ├── org.v2mom.json     # Organization V2MOM
│   └── teams/             # Team V2MOMs
├── capability/
│   ├── stack.capability.json
│   └── domains/           # Domain capabilities
├── roadmap/
│   └── roadmap.json       # Initiative timeline
└── maturity/
    ├── model.json         # Maturity model definition
    └── state.json         # Current assessment
```

## Creating Your Own Samples

You can create custom samples:

1. Set up a project with desired structure
2. Place in `samples/` directory
3. Add `project.json` with metadata
4. Add `README.md` describing the sample
5. Sample appears in sample picker

### Sample Metadata

```json
{
  "metadata": {
    "id": "my-sample",
    "title": "My Custom Sample",
    "description": "Example for specific use case",
    "version": "1.0.0",
    "author": "Your Name"
  }
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/samples` | GET | List available samples |
| `/api/samples/{id}` | GET | Get sample details |
| `/api/samples/{id}/import` | POST | Import sample to workspace |

## Related

- [Projects](projects.md)
- [V2MOM Cascade](v2mom.md)
- [Capabilities](capabilities.md)
