# System Architecture Overview Guide

## Overview
The AIO Fund Manager now includes a comprehensive **System Architecture Overview** page that visualizes the entire platform structure, showing both the Operations Portal and Client Portal with their interconnected modules.

## Accessing System Overview

### Method 1: Profile Menu
1. Click on your profile avatar (JD) in the top-right corner
2. Select **"System Overview"** from the dropdown menu
3. You'll be taken to the full system architecture visualization

### Method 2: Direct URL
- Navigate to `/system-overview` route

## Features

### 1. **Three Major System Groups**

#### Operations Portal (Dark Navy)
Internal admin and staff management portal with 9 core modules:
- **CRM (Leads + Clients)** - Manage leads and client relationships
- **Trust Registry** - Central trust record database
- **Document Center** - Document management and storage
- **Compliance Tracker** - Track compliance requirements
- **Scheduling** - Appointment and meeting management
- **Messaging** - Multi-channel communication (Chat/SMS/Email)
- **Template Builder** - Create document templates
- **Reporting + Capital Log** - Analytics and financial tracking
- **Settings + Roles** - System configuration and permissions

#### Client Portal (Gold)
Customer-facing interface with 7 core modules:
- **Onboarding Status** - Client activation progress
- **Document Upload** - Client document submission
- **Education Center** - Learning resources and AI assistant
- **Trust Assets View** - View trust portfolio
- **Fund Marketplace** - Browse investment opportunities (post-verification)
- **Schedule Calls** - Book consultant appointments
- **Secure Messages** - Client-consultant communication

#### Shared Services (Gray)
Cross-platform infrastructure services:
- **Authentication + Role Routing** - User authentication and access control
- **Notifications + Email/SMS** - Multi-channel notifications
- **File Storage + Tagging** - Secure file management
- **eSignature Integration** - Digital signature workflows

### 2. **Interactive Module Cards**
- Click any module to view detailed information
- See module type (Feature Module vs Shared Service)
- View all connections related to the selected module
- Highlighted connections show data flow and integration points

### 3. **Connection Visualization**
The bottom section shows all system connections:
- Visual representation of data flow between modules
- Labeled arrows indicating the type of data/process being shared
- Interactive highlighting when selecting modules
- Color-coded to show source and destination

### 4. **Connection Types**
Current system connections include:
- **CRM → Trust Registry**: Create Client → Trust Record
- **Document Center → Trust Registry**: Trust Docs
- **Compliance Tracker → Onboarding**: Show Status to Client
- **Scheduling → Schedule Calls**: Sync Appointments
- **Messaging → Secure Messages**: Two-way Messaging
- **CRM → Onboarding**: Lead → Verified Client

## Visual Design

### Color Scheme
- **Operations Portal**: Dark Navy (#0B1930) - Professional, authoritative
- **Client Portal**: Gold (#C6A661) - Premium, welcoming
- **Shared Services**: Slate Gray (#64748B) - Neutral, foundational
- **Highlights**: Gold (#C6A661) for active connections

### Interactive States
- **Default**: Light border, white background
- **Selected**: Bold colored border, colored background tint, shadow
- **Connected**: Gold border and tinted background when related to selected module
- **Hover**: Subtle border and background change

## Use Cases

### 1. **System Planning**
Use this view to:
- Understand the full platform architecture
- Plan new features and integrations
- Identify dependencies between modules
- Map the Nexxess 5-Step Trust Process to existing modules

### 2. **Developer Onboarding**
Help new developers understand:
- How the system is organized
- Which modules communicate with each other
- The difference between Operations and Client portals
- Shared services used across the platform

### 3. **Business Analysis**
Enable stakeholders to:
- See the complete system at a glance
- Understand client vs operations workflows
- Identify integration points for new vendors
- Document system capabilities

### 4. **Process Mapping**
Prepare for mapping new processes like:
- Nexxess 5-Step Trust Book process
- Custom client onboarding workflows
- Compliance tracking procedures
- Document management protocols

## Technical Implementation

### Components
- **SystemOverviewPage** (`/components/pages/SystemOverviewPage.tsx`)
  - Main visualization component
  - Interactive module selection
  - Connection highlighting logic
  - Responsive grid layout

### Routing
- Route: `/system-overview`
- Type definition added to `Route` type in `App.tsx`
- Integrated into `renderPage()` switch statement
- Page title: "System Overview"

### Navigation
- Added to profile dropdown menu in `DashboardLayout.tsx`
- Network icon for visual identification
- Accessible from any authenticated page

## Future Enhancements

### Planned Features
1. **Drag-and-Drop Layout** - Rearrange modules visually
2. **Module Details Drawer** - Deep dive into each module's features
3. **Connection Animation** - Animated data flow visualization
4. **Export Diagram** - Download as PNG/PDF for documentation
5. **Search & Filter** - Find specific modules or connections
6. **Version History** - Track system architecture changes over time
7. **Process Overlay** - Overlay specific business processes (like Nexxess 5-Step)

### Integration Opportunities
- **API Documentation** - Link to API docs for each module
- **Status Monitoring** - Show real-time module health
- **Usage Analytics** - Display usage stats per module
- **Dependency Graph** - Advanced connection visualization

## Notes

### Implementation Notes
- All modules are currently implemented in the platform
- This visualization serves as documentation and planning tool
- The system supports both internal staff and external client users
- Role-based access control routes users to appropriate portals

### For Nexxess 5-Step Trust Process
This architecture overview will serve as the foundation for mapping the Nexxess 5-Step Trust Book process:
1. Use this view to identify which existing modules support each step
2. Determine where new functionality needs to be added
3. Plan integration points between modules
4. Document the end-to-end workflow across both portals

## Support

For questions or issues with the System Overview page:
- Check this guide for features and usage
- Review module descriptions for clarification
- Use the interactive features to explore connections
- Reference this for system architecture discussions

---

**Last Updated**: December 11, 2025  
**Version**: 1.0  
**Status**: ✅ Fully Implemented
