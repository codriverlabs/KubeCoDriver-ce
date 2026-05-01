# KubeCoDriver Roadmap

This document outlines the planned features and improvements for the KubeCoDriver (Temporary Observability Engine) project.

## CoDriverJob Schema Enhancements

### 1. Ephemeral Container Attachment
- **Goal**: Enable CoDriverJob to attach ephemeral containers at the start of selected containers from pods
- **Approach**: Enhance CoDriverJob schema to support container-level targeting and ephemeral container specifications
- **Priority**: High
- **Status**: Planned

### 2. Command Line Arguments Field
- **Goal**: Add commandLineArguments field to CoDriverJob schema
- **Approach**: Allow end users to specify extra command line arguments for profiling tools
- **Priority**: High
- **Status**: Planned

## Fault Injection Capabilities

### 3. Storage Fault Injection
- **Goal**: Implement PV (Persistent Volume) fill-up fault injection
- **Approach**: Create CoDriverJob examples that can fill up storage to test application resilience
- **Priority**: Medium
- **Status**: Planned

### 4. CPU Usage Fault Injection
- **Goal**: Implement CPU stress fault injection
- **Approach**: Create CoDriverJob examples that can raise CPU usage in target pods
- **Priority**: Medium
- **Status**: Planned

### 5. Memory Usage Fault Injection
- **Goal**: Implement memory stress fault injection
- **Approach**: Create CoDriverJob examples that can raise memory usage in target pods
- **Priority**: Medium
- **Status**: Planned

### 6. Process Termination Fault Injection
- **Goal**: Implement process termination capabilities
- **Features**:
  - Abrupt process termination (SIGKILL)
  - Graceful process termination (SIGTERM)
  - Target main container processes
  - Target sidecar container processes
- **Priority**: Medium
- **Status**: Planned

## Security & Safety Enhancements

### 7. Command Line Argument Sanitization
- **Goal**: Implement extra sanitization for command line arguments in CoDriverJob
- **Approach**: Define sub-CRD that will limit the vector of attack
- **Priority**: High
- **Status**: Planned

### 8. CoDriverTool Argument Enforcement
- **Goal**: Enforce acceptable command line arguments in CoDriverTool
- **Approach**: Implement validation rules and whitelisting mechanisms
- **Priority**: High
- **Status**: Planned

## Reliability & Resource Management

### 9. Signal Handling for CoDriverJob
- **Goal**: Add handling and trapping termination signals for each CoDriverJob
- **Approach**: Implement graceful shutdown to save work when collection pod gets interrupted
- **Priority**: Medium
- **Status**: Planned

### 10. Default Resource Specifications
- **Goal**: CoDriverTool should specify default CPU and memory requests and limits
- **Approach**: Define resource templates and defaults in configuration
- **Priority**: Medium
- **Status**: Planned

## Data Processing & Visualization

### 11. CoDriverJobProcessing CRD
- **Goal**: Define CoDriverJobProcessingCRD for each collector
- **Features**:
  - Specify frequency of processing collected data
  - Define how to expose processed data
  - For aperf: HTML reports exposed via ingress
  - Processing using LLMs/SLMs for automated RCA (Root Cause Analysis)
- **Priority**: High
- **Status**: Planned

## Development & Deployment Optimization

### 12. Helm Charts Optimization
- **Goal**: Optimize helm charts generation
- **Approach**: Streamline chart structure and improve templating
- **Priority**: Medium
- **Status**: Planned

### 13. GitHub Actions Artifacts
- **Goal**: Optimize generation of artifacts via GitHub Actions
- **Approach**: Improve CI/CD pipeline efficiency and artifact management
- **Priority**: Medium
- **Status**: Planned

## User Experience & Adoption

### 14. Interactive Tutorial
- **Goal**: Implement "Hello CoDriverJob" interactive tutorial
- **Approach**: Create step-by-step guided experience for new users
- **Priority**: Low
- **Status**: Planned

### 15. Community Engagement
- **Goal**: Have lots of workshops and coffees
- **Approach**: Organize community events, workshops, and informal meetups
- **Priority**: Ongoing
- **Status**: Planned

## Implementation Timeline

- **Phase 1**: CoDriverJob Schema Enhancements (Items 1-2)
- **Phase 2**: Security & Safety (Items 7-8)
- **Phase 3**: Fault Injection Capabilities (Items 3-6)
- **Phase 4**: Reliability & Resources (Items 9-10)
- **Phase 5**: Data Processing (Item 11)
- **Phase 6**: Optimization (Items 12-13)
- **Phase 7**: User Experience (Items 14-15)

## Contributing

Contributions to any of these roadmap items are welcome. Please see our contributing guidelines and feel free to open issues or pull requests for discussion.
