# Kubectl Kanvas Snapshot Implementation

This document outlines the implementation details and workflow for the Kubectl Kanvas Snapshot plugin.

## Overview

Kubectl Kanvas Snapshot is a Kubectl plugin that creates visual snapshots of Kubernetes resources using Meshery's visualization capabilities. It transforms Kubernetes manifests into interactive designs that can be viewed, shared, and analyzed.

## Workflow Diagram

The following diagram illustrates the end-to-end workflow of the plugin:

```mermaid
graph TD;
    START([Start Kubectl Kanvas Snapshot])-->INPUT{Input Parameters};
    INPUT-->|Email Provided|EMAIL_VALIDATION{"Validate Email"};
    INPUT-->|No Email|AUTH{"Read Auth Credentials"};
    EMAIL_VALIDATION-->|Valid Email|EMAIL_PROCESS["Process Email Notification"];
    EMAIL_VALIDATION-->|Invalid Email|EMAIL_ERROR["Display Email Validation Error"];
    EMAIL_ERROR-->EXIT([Exit Process]);
    EMAIL_PROCESS-->AUTH;
    AUTH-->MESHERY_REQ["Send Request to /api/pattern/import"];
    MESHERY_REQ-->RESOURCE_PROCESS{"Resources Processable?"};
    RESOURCE_PROCESS-->|Yes|DESIGN_GEN["Generate Design Visualization"];
    RESOURCE_PROCESS-->|No|RESOURCE_ERROR["Display Resource Processing Error"];
    RESOURCE_ERROR-->FINAL_EXIT([Exit Process]);
    DESIGN_GEN-->DESIGN_ID["Retrieve Design ID"];
    DESIGN_ID-->GITHUB_PREP["Prepare kanvas.yaml Workflow"];
    GITHUB_PREP-->WORKFLOW_TRIGGER["Trigger layer5labs/kanvas-snapshot"];
    WORKFLOW_TRIGGER-->CYPRESS_SETUP["Setup Cypress Environment"];
    CYPRESS_SETUP-->BROWSER_LAUNCH["Launch Headless Browser"];
    BROWSER_LAUNCH-->MESHERY_NAV["Navigate to Meshery UI"];
    MESHERY_NAV-->RENDER_WAIT["Wait for Design Rendering"];
    RENDER_WAIT-->SCREENSHOT["Capture Screenshot"];
    SCREENSHOT-->|Email Provided|EMAIL_NOTIFY["Send Email Notification"];
    SCREENSHOT-->|No Email|LINK_DISPLAY["Display Snapshot Links"];
    EMAIL_NOTIFY-->LINK_DISPLAY;
    LINK_DISPLAY-->GITHUB_LINK["Show GitHub Snapshot Link"];
    GITHUB_LINK-->END([Workflow Complete]);
```

## Implementation Details

The workflow consists of the following key steps:

1. **Input Processing**:
   - Parse Kubernetes manifest files
   - Validate email if provided
   - Check authentication credentials

2. **Meshery Integration**:
   - Send base64-encoded manifest to Meshery's API
   - Process the response to extract the design ID

3. **GitHub Workflow**:
   - Trigger the GitHub Actions workflow in the layer5labs/kanvas-snapshot repository
   - Pass design ID and other parameters

4. **Snapshot Generation**:
   - Cypress browser automation captures screenshots of the design
   - Screenshots are stored as artifacts in the GitHub workflow run

5. **Result Communication**:
   - Display links to view the design in Meshery
   - Show where to find the generated screenshots
   - Send email notification if an email was provided

