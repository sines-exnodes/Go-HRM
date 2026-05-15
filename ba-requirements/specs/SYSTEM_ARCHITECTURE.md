# AIO Fund Manager - System Architecture

## Overview
AIO Fund Manager is a white-label SaaS platform that provides a dual-portal system for financial services trust activation and fund management processes. The platform consists of an Operations Portal (internal staff) and a Client Portal (external clients).

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CLIENT LAYER                            │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  Client Portal   │         │ Operations Portal │         │
│  │  (React + TS)    │         │  (React + TS)     │         │
│  └──────────────────┘         └──────────────────┘         │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                  PRESENTATION LAYER                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Branding Context (White-Label Theming)              │  │
│  │  • Logo, Colors, Company Info                        │  │
│  │  • Login Info Cards, Learning Videos                 │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   APPLICATION LAYER                         │
│  ┌────────────┐  ┌────────────┐  ┌─────────────┐          │
│  │ Workflow   │  │ Compliance │  │  Document   │          │
│  │ Engine     │  │ Manager    │  │  Manager    │          │
│  └────────────┘  └────────────┘  └─────────────┘          │
│  ┌────────────┐  ┌────────────┐  ┌─────────────┐          │
│  │ Messaging  │  │ Calendar   │  │  AI         │          │
│  │ System     │  │ System     │  │  Assistant  │          │
│  └────────────┘  └────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      DATA LAYER                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Data Models                                         │  │
│  │  • Portfolios (Trusts, PE Funds, RE Funds)          │  │
│  │  • Compliance Tasks & Workflows                     │  │
│  │  • Documents, Messages, Appointments                │  │
│  │  • Users, Notifications, Activity Timeline          │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   INFRASTRUCTURE LAYER                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Supabase (Optional)                                 │  │
│  │  • Authentication, Database, Storage, Real-time      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Portal Architecture

### 1. Client Portal
**Purpose:** External-facing portal for clients to manage their portfolios and complete compliance workflows.

**Key Features:**
- Portfolio Dashboard
- Compliance Task Management
- Document Upload & Management
- Messaging with Advisors
- Calendar & Appointments
- Learning Center
- AI Assistant

**User Roles:**
- Client (Authenticated)
- Guest (Limited Access)

### 2. Operations Portal
**Purpose:** Internal portal for financial advisors to configure workflows and manage client processes.

**Key Features:**
- Client Portal Settings (Branding, Login, Learning Center)
- Workflow Configuration
- Client Management
- Compliance Monitoring
- Analytics & Reporting

**User Roles:**
- Admin
- Financial Advisor
- Compliance Officer

---

## Technology Stack

### Frontend
- **Framework:** React 18+ with TypeScript
- **Styling:** Tailwind CSS v4
- **UI Components:** shadcn/ui
- **State Management:** React Context API
- **Routing:** Client-side routing
- **Icons:** Lucide React
- **Charts:** Recharts
- **Date Handling:** date-fns
- **Notifications:** Sonner

### Backend (Optional - Supabase)
- **Authentication:** Supabase Auth
- **Database:** PostgreSQL (via Supabase)
- **Storage:** Supabase Storage
- **Real-time:** Supabase Realtime

---

## Core Components

### 1. Branding System
**Location:** `/contexts/BrandingContext.tsx`

**Functionality:**
- Centralized branding configuration
- CSS variable injection for dynamic theming
- White-label customization support

**Configurable Elements:**
- Logo & Company Info
- Color Scheme (Primary, Secondary, Accent)
- Login Info Cards
- Learning Videos
- Social Media Links
- AI Assistant Toggle

### 2. Workflow Engine
**Location:** `/data/workflowData.ts`

**Supported Workflows:**
1. **Trust - Nexxess Trust** (Blue) - 22 Steps
2. **Fund - Private Equity** (Purple) - 22 Steps
3. **Fund - Real Estate** (Green) - 22 Steps

**Task Types:**
- Upload (Document submission)
- Review (Agreement review & signature)
- Schedule (Appointment booking)
- Simple (Information display)

### 3. Compliance Manager
**Location:** Various panel components in `/components/panels/`

**Features:**
- Dynamic form generation
- Multi-step workflows
- Data validation
- Progress tracking
- Status management (Not Started, In Progress, Completed, Under Review)

### 4. Document Management
**Location:** `/components/pages/TrustDetailPage.tsx` (Documents Tab)

**Features:**
- Document upload & download
- Status tracking (Uploaded, Pending, Approved)
- Category organization
- Version control support

### 5. Messaging System
**Location:** `/components/pages/MessagingPage.tsx`

**Features:**
- Portfolio-specific conversations
- Real-time message threads
- Advisor-client communication
- Attachment support

### 6. Calendar System
**Location:** `/components/pages/CalendarPage.tsx`

**Features:**
- Appointment scheduling
- Portfolio-linked events
- Multiple view modes (Month, Week, Day)
- Status management (Scheduled, Completed, Cancelled)

### 7. AI Assistant
**Location:** `/components/AIAssistant.tsx`

**Features:**
- Contextual help
- Document assistance
- FAQ responses
- Integration with Learning Center

---

## Data Models

### Portfolio
```typescript
interface Portfolio {
  id: string;
  name: string;
  type: 'trust' | 'fund-pe' | 'fund-re';
  workflow: string;
  status: 'active' | 'pending' | 'completed';
  progress: number;
  totalValue?: number;
  lastActivity: string;
}
```

### Compliance Task
```typescript
interface ComplianceTask {
  id: string;
  taskId: string;
  label: string;
  type: 'upload' | 'review' | 'schedule' | 'simple';
  status: 'not-started' | 'in-progress' | 'completed' | 'under-review';
  dueDate?: string;
  completedDate?: string;
  formData?: Record<string, any>;
}
```

### Document
```typescript
interface Document {
  id: string;
  name: string;
  category: string;
  uploadDate: string;
  status: 'uploaded' | 'pending' | 'approved';
  size: string;
  type: string;
}
```

### Message
```typescript
interface Message {
  id: string;
  sender: string;
  senderRole: 'client' | 'advisor';
  content: string;
  timestamp: string;
  read: boolean;
}
```

### Appointment
```typescript
interface Appointment {
  id: string;
  title: string;
  date: Date;
  time: string;
  duration: string;
  type: string;
  status: 'scheduled' | 'completed' | 'cancelled';
  portfolioId: string;
  portfolioName: string;
  advisor?: string;
  location?: string;
  notes?: string;
}
```

---

## Security & Compliance

### Authentication
- Email/Password authentication
- Guest access (limited features)
- Session management
- Logout functionality

### Data Protection
- Client-side validation
- Sensitive data handling
- Document encryption (when using Supabase)
- Access control by user role

### Compliance Features
- KYC/AML workflow support
- Document retention
- Audit trail (Activity Timeline)
- Status tracking for regulatory requirements

---

## Performance Optimization

### Code Splitting
- Route-based code splitting
- Lazy loading for modals and heavy components

### State Management
- Context API for global state
- Local state for component-specific data
- Minimal re-renders through proper component structure

### Asset Optimization
- Image lazy loading (ImageWithFallback component)
- Unsplash integration for dynamic images
- SVG optimization for icons

---

## Deployment Architecture

### Frontend Deployment
```
┌─────────────────────┐
│   CDN / Static Host │
│   (Vercel, Netlify) │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│   React Application │
│   (Client Portal)   │
└─────────────────────┘
```

### Full-Stack Deployment (with Supabase)
```
┌─────────────────────┐
│   CDN / Static Host │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│   React Application │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│   Supabase Backend  │
│   • Auth            │
│   • Database        │
│   • Storage         │
│   • Real-time       │
└─────────────────────┘
```

---

## Integration Points

### External Services
- **Unsplash API:** Stock images for UI
- **Supabase (Optional):** Backend services
- **AI Services (Future):** Advanced AI assistant features

### White-Label Customization
- Branding configuration via Operations Portal
- CSS variable injection for theming
- Logo and company info customization
- Learning video library management

---

## Scalability Considerations

### Horizontal Scaling
- Stateless frontend architecture
- CDN distribution for static assets
- Database connection pooling (Supabase)

### Vertical Scaling
- Optimized React rendering
- Efficient data structures
- Memoization for expensive computations

### Data Growth
- Pagination for large datasets
- Lazy loading for documents and messages
- Archive/retention policies for old data

---

## Monitoring & Analytics

### Client Portal Metrics
- User engagement (page views, time spent)
- Task completion rates
- Document upload/approval times
- Message response times

### Operations Portal Metrics
- Workflow configuration usage
- Client onboarding times
- Compliance task status distribution
- System performance metrics

---

## Future Enhancements

### Planned Features
1. Advanced AI Assistant with NLP
2. E-signature integration
3. Payment processing
4. Multi-language support
5. Mobile native apps
6. Advanced analytics dashboard
7. Automated compliance reporting
8. Integration with CRM systems

### Technical Improvements
1. GraphQL API layer
2. Advanced caching strategies
3. Service worker for offline support
4. Real-time collaboration features
5. Advanced security features (2FA, SSO)

---

## Version History

- **v1.0** - Initial release with basic portfolio management
- **v1.1** - Added messaging and calendar features
- **v1.2** - Implemented configurable branding and Learning Center
- **Current** - Full dual-portal system with 3 complete workflows

---

## Support & Documentation

### Developer Documentation
- `/SYSTEM_OVERVIEW_GUIDE.md` - Complete system overview
- `/CLIENT_FORMS_GUIDE.md` - Form panel implementation
- `/COMPLIANCE_WORKFLOW_GUIDE.md` - Workflow configuration
- `/WHITE_LABEL_GUIDE.md` - Branding customization

### User Documentation
- Learning Center videos
- AI Assistant contextual help
- In-app tooltips and guidance

---

*Last Updated: January 2026*
*Version: 1.2*
*Platform: AIO Fund Manager - Nexxess Business Advisors*
