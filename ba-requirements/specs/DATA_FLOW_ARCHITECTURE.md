# AIO Fund Manager - Data Flow Architecture

## Overview
This document details the data flow patterns, state management, and data synchronization across the AIO Fund Manager platform.

---

## 1. Client Portal Data Flow

### Authentication Flow
```
┌─────────────────────────────────────────────────────────────┐
│                    Login Page                               │
│                    • Email Input                            │
│                    • Password Input                         │
│                    • Guest Access Button                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Submit      │
                  │  Credentials │
                  └──────────────┘
                          │
                ┌─────────┴─────────┐
                │                   │
               Mock              Supabase
             (Current)           (Optional)
                │                   │
                ▼                   ▼
    ┌──────────────────┐  ┌──────────────────┐
    │  Validate Email  │  │  supabase.auth   │
    │  Auto-Accept     │  │  .signIn()       │
    └──────────────────┘  └──────────────────┘
                │                   │
                └─────────┬─────────┘
                          │
                          ▼
                ┌──────────────────┐
                │  Set Auth State  │
                │  isAuthenticated │
                │  = true          │
                └──────────────────┘
                          │
                          ▼
                ┌──────────────────┐
                │  Load User Data  │
                │  • Profile       │
                │  • Portfolios    │
                │  • Notifications │
                └──────────────────┘
                          │
                          ▼
                ┌──────────────────┐
                │  Redirect to     │
                │  Dashboard       │
                └──────────────────┘
```

### Portfolio Data Loading
```
┌─────────────────────────────────────────────────────────────┐
│                    App Component Mount                      │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Load Portfolio Data                      │
│                    Source: /data/trustData.ts               │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Parse Portfolio Definitions                    │
│              • Johnson Family Trust (ID: 1)                 │
│              • Venture Growth Fund LP (ID: 2)               │
│              • Metro Property Fund (ID: 3)                  │
│              • Anderson Estate Trust (ID: 4)                │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Load Associated Workflows                      │
│              Source: /data/workflowData.ts                  │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Nexxess     │ │  Private     │ │  Real Estate │
    │  Trust       │ │  Equity      │ │  Fund        │
    │  (22 steps)  │ │  (22 steps)  │ │  (22 steps)  │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              └───────────┴───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Initialize Compliance State                    │
│              complianceTasks: Record<string, Task[]>        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Calculate Initial Progress                     │
│              • Count completed tasks                        │
│              • Update progress percentage                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Render Dashboard with Data                     │
└─────────────────────────────────────────────────────────────┘
```

### Task Completion Flow
```
┌─────────────────────────────────────────────────────────────┐
│              User Clicks Task in Checklist                  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Open Task Panel (Drawer)                       │
│              • Load task definition                         │
│              • Check current status                         │
│              • Load existing formData (if any)              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              User Fills Form / Completes Action             │
│              • Upload documents                             │
│              • Enter form fields                            │
│              • Review agreements                            │
│              • Schedule appointments                        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Click "Submit" or "Complete"                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │  Validation  │
                  │  Required?   │
                  └──────────────┘
                    │          │
                   Yes        No
                    │          │
                    ▼          │
          ┌──────────────┐    │
          │  Validate    │    │
          │  Form Data   │    │
          └──────────────┘    │
                │   │          │
              Valid Invalid    │
                │   │          │
                ▼   ▼          │
                │  Show Error  │
                │   │          │
                └───┘          │
                    │          │
                    └──────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              handleTaskComplete(taskId, formData)           │
│              Called from panel component                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update State in App.tsx                        │
│              setComplianceTasks((prev) => ({                │
│                ...prev,                                     │
│                [portfolioId]: updatedTasks                  │
│              }))                                            │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update Task Properties                         │
│              • status: 'completed'                          │
│              • completedDate: new Date()                    │
│              • formData: {...submittedData}                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Side Effects (Parallel)                        │
└─────────────────────────────────────────────────────────────┘
              │           │           │           │
              ▼           ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Add to      │ │  Create      │ │  Recalculate │ │  Show Toast  │
    │  Timeline    │ │  Notification│ │  Progress    │ │  Success Msg │
    └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │           │
              └───────────┴───────────┴───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              UI Re-renders with New State                   │
│              • Checklist shows green checkmark              │
│              • Progress bar updates                         │
│              • Timeline shows new entry                     │
│              • Notification badge updates                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Close Task Panel (Drawer)                      │
└─────────────────────────────────────────────────────────────┘
```

### Document Upload Data Flow
```
┌─────────────────────────────────────────────────────────────┐
│              User Selects File                              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              File Input onChange Event                      │
│              const file = event.target.files[0]             │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Client-Side Validation                         │
│              • Check file type                              │
│              • Check file size                              │
│              • Validate name                                │
└─────────────────────────────────────────────────────────────┘
                          │
                ┌─────────┴─────────┐
                │                   │
              Valid            Invalid
                │                   │
                ▼                   ▼
                │          ┌──────────────┐
                │          │  Show Error  │
                │          │  Message     │
                │          └──────────────┘
                │                   │
                │                   ▼
                │          ┌──────────────┐
                │          │  Stop        │
                │          │  Process     │
                │          └──────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│              Show Upload Progress UI                        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                ┌──────────────────┐
                │  Backend Type?   │
                └──────────────────┘
                │              │
              Mock         Supabase
                │              │
                ▼              ▼
    ┌──────────────────┐  ┌──────────────────┐
    │  Simulate Upload │  │  Upload to       │
    │  Create FileURL  │  │  Storage Bucket  │
    │  (base64/mock)   │  │                  │
    └──────────────────┘  └──────────────────┘
                │              │
                └──────┬───────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Create Document Record                         │
│              {                                              │
│                id: uuid(),                                  │
│                name: file.name,                             │
│                category: selectedCategory,                  │
│                uploadDate: new Date(),                      │
│                status: 'uploaded',                          │
│                size: file.size,                             │
│                type: file.type,                             │
│                url: uploadedUrl                             │
│              }                                              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Save to Documents Array                        │
│              Update Portfolio State                         │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update UI                                      │
│              • Show document in list                        │
│              • Update document count                        │
│              • Add to timeline                              │
│              • Show success toast                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. State Management Architecture

### Global State (App.tsx)
```typescript
// Authentication State
const [isAuthenticated, setIsAuthenticated] = useState(false);
const [userType, setUserType] = useState<'client' | 'guest' | null>(null);

// Navigation State
const [currentRoute, setCurrentRoute] = useState<Route>('/dashboard');

// Portfolio State
const [selectedTrustId, setSelectedTrustId] = useState<string | null>(null);
const [selectedFundId, setSelectedFundId] = useState<string | null>(null);

// Compliance Tasks State
const [complianceTasks, setComplianceTasks] = useState<
  Record<string, ComplianceTask[]>
>({
  '1': [...nexxessTrustTasks],
  '2': [...ventureGrowthTasks],
  '3': [...metroPropertyTasks],
  '4': [...andersonEstateTasks],
});

// Notifications State
const [notifications, setNotifications] = useState<Notification[]>([]);

// Modal State
const [uploadModalOpen, setUploadModalOpen] = useState(false);
const [scheduleModalOpen, setScheduleModalOpen] = useState(false);
const [reviewModalOpen, setReviewModalOpen] = useState(false);
const [videoModalOpen, setVideoModalOpen] = useState(false);

// Appointment State
const [appointments, setAppointments] = useState<Appointment[]>([]);
const [selectedAppointment, setSelectedAppointment] = useState<Appointment | null>(null);
```

### Context-Based State

#### BrandingContext
```typescript
// Global Branding Configuration
interface ClientPortalSettings {
  logo: string;
  primaryColor: string;
  secondaryColor: string;
  accentColor: string;
  loginInfoCards: LoginInfoCard[];
  learningVideos: LearningVideo[];
  aiAssistantEnabled: boolean;
  socialLinks: SocialLink[];
  companyName: string;
  companyInitials: string;
  showPoweredBy: boolean;
}

// Provider wraps entire app
<BrandingProvider>
  <App />
</BrandingProvider>

// Components access via hook
const branding = useBranding();
```

### Local Component State
```typescript
// Example: SchedulingTaskPanel
const [selectedDate, setSelectedDate] = useState<Date | undefined>();
const [selectedTime, setSelectedTime] = useState<string>('');
const [meetingNotes, setMeetingNotes] = useState<string>('');

// Example: UploadTaskPanel
const [selectedDocument, setSelectedDocument] = useState<string>('');
const [uploadNotes, setUploadNotes] = useState<string>('');
const [file, setFile] = useState<File | null>(null);
```

---

## 3. Data Persistence Patterns

### Mock Data Pattern (Current)
```
┌─────────────────────────────────────────────────────────────┐
│                    Static Data Files                        │
│                    /data/trustData.ts                       │
│                    /data/workflowData.ts                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    App.tsx State                            │
│                    useState() hooks                         │
│                    In-memory storage                        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Lifecycle                           │
│                    • Persists during session                │
│                    • Lost on page refresh                   │
│                    • Reset on logout                        │
└─────────────────────────────────────────────────────────────┘
```

### Supabase Pattern (Optional)
```
┌─────────────────────────────────────────────────────────────┐
│                    Supabase Database                        │
│                    PostgreSQL Tables                        │
│                    • portfolios                             │
│                    • compliance_tasks                       │
│                    • documents                              │
│                    • messages                               │
│                    • appointments                           │
│                    • notifications                          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Supabase Client                          │
│                    supabase.from('table')                   │
│                    .select() / .insert() / .update()        │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    React State Sync                         │
│                    • Initial load from DB                   │
│                    • Real-time subscriptions                │
│                    • Optimistic updates                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    UI Rendering                             │
│                    • Always shows latest data               │
│                    • Persists across sessions               │
│                    • Multi-device sync                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. Event-Driven Data Flow

### Task Completion Event Chain
```
User Action
    │
    ▼
┌──────────────────────────────────┐
│  handleTaskComplete(id, data)   │
└──────────────────────────────────┘
    │
    ├─► Update complianceTasks state
    │
    ├─► Add timeline entry
    │   └─► setTimelineData((prev) => [...prev, newEntry])
    │
    ├─► Create notification
    │   └─► setNotifications((prev) => [...prev, newNotif])
    │
    ├─► Recalculate progress
    │   └─► Calculate completed / total ratio
    │
    ├─► Show toast notification
    │   └─► toast.success("Task completed!")
    │
    └─► Close task panel
        └─► setDrawerOpen(false)
```

### Appointment Creation Event Chain
```
User Schedules Appointment
    │
    ▼
┌──────────────────────────────────┐
│  handleScheduleAppointment()    │
└──────────────────────────────────┘
    │
    ├─► Create appointment object
    │   └─► { id, title, date, time, status, portfolioId, ... }
    │
    ├─► Add to appointments array
    │   └─► setAppointments((prev) => [...prev, newAppt])
    │
    ├─► Update calendar view
    │   └─► CalendarPage re-renders with new event
    │
    ├─► Add to timeline
    │   └─► New activity entry "Appointment scheduled"
    │
    ├─► Create notification
    │   └─► "Appointment confirmed for {date}"
    │
    ├─► Update task status
    │   └─► Mark scheduling task as complete
    │
    └─► Show confirmation toast
        └─► toast.success("Appointment scheduled!")
```

### Document Upload Event Chain
```
User Uploads Document
    │
    ▼
┌──────────────────────────────────┐
│  handleDocumentUpload(file)     │
└──────────────────────────────────┘
    │
    ├─► Validate file
    │   └─► Check type, size, name
    │
    ├─► Upload to storage
    │   └─► Mock: create data URL
    │   └─► Supabase: upload to bucket
    │
    ├─► Create document record
    │   └─► { id, name, category, date, status, ... }
    │
    ├─► Add to documents array
    │   └─► Update portfolio documents list
    │
    ├─► Add to timeline
    │   └─► "Document uploaded: {name}"
    │
    ├─► Create notification (for advisor)
    │   └─► "New document requires review"
    │
    ├─► Update task status (if applicable)
    │   └─► Mark upload task as complete
    │
    └─► Show success toast
        └─► toast.success("Document uploaded successfully!")
```

---

## 5. Messaging Data Flow

### Real-Time Messaging Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                    Client Portal UI                         │
│                    MessagingPage Component                  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Message State Management                       │
│              conversations: Conversation[]                  │
│              selectedConversation: string | null            │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┴───────────┐
              │                       │
              ▼                       ▼
    ┌──────────────────┐    ┌──────────────────┐
    │  Send Message    │    │  Receive Message │
    └──────────────────┘    └──────────────────┘
              │                       │
              ▼                       │
    ┌──────────────────┐             │
    │  Create Message  │             │
    │  Object          │             │
    │  {               │             │
    │    id: uuid(),   │             │
    │    sender,       │             │
    │    content,      │             │
    │    timestamp,    │             │
    │    read: false   │             │
    │  }               │             │
    └──────────────────┘             │
              │                       │
              ▼                       │
    ┌──────────────────┐             │
    │  Add to Thread   │             │
    └──────────────────┘             │
              │                       │
              └───────────┬───────────┘
                          │
                          ▼
              ┌──────────────────────┐
              │  Update UI           │
              │  • Message list      │
              │  • Unread count      │
              │  • Last message time │
              └──────────────────────┘
                          │
                          ▼
              ┌──────────────────────┐
              │  Create Notification │
              │  (for recipient)     │
              └──────────────────────┘
```

### Message Thread Structure
```
Conversation {
  id: string
  portfolioId: string
  portfolioName: string
  advisor: {
    name: string
    avatar: string
    role: string
  }
  lastMessage: string
  lastMessageTime: string
  unreadCount: number
  messages: Message[]
}

Message {
  id: string
  sender: string
  senderRole: 'client' | 'advisor'
  content: string
  timestamp: string
  read: boolean
  attachments?: Attachment[]
}
```

---

## 6. Calendar & Appointment Data Flow

### Appointment State Management
```
┌─────────────────────────────────────────────────────────────┐
│                    Appointments Array                       │
│                    useState<Appointment[]>([...])           │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┴───────────┐
              │                       │
              ▼                       ▼
    ┌──────────────────┐    ┌──────────────────┐
    │  CalendarPage    │    │  Dashboard       │
    │  • Month View    │    │  • Next Appt     │
    │  • Week View     │    │  • Upcoming      │
    │  • Day View      │    │                  │
    └──────────────────┘    └──────────────────┘
              │                       │
              └───────────┬───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Appointment Actions                            │
└─────────────────────────────────────────────────────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  View        │ │  Reschedule  │ │  Cancel      │
    │  Details     │ │              │ │              │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Open Modal  │ │  Update Date │ │  Set Status  │
    │              │ │  & Time      │ │  'cancelled' │
    └──────────────┘ └──────────────┘ └──────────────┘
                          │
                          ▼
              ┌──────────────────────┐
              │  Update State        │
              │  Create Timeline     │
              │  Send Notification   │
              └──────────────────────┘
```

### Calendar Date Grouping
```
appointments[]
    │
    ▼
Group by Date
    │
    ├─► 2026-01-15: [appt1, appt2]
    ├─► 2026-01-18: [appt3]
    └─► 2026-01-22: [appt4, appt5, appt6]
    │
    ▼
Render in Calendar Grid
    │
    ├─► Month View: Show dots for dates with appointments
    ├─► Week View: Show appointment blocks
    └─► Day View: Show detailed schedule
```

---

## 7. Notification System Data Flow

### Notification Creation & Distribution
```
┌─────────────────────────────────────────────────────────────┐
│                    System Event Occurs                      │
│                    (Task complete, Document upload, etc.)   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Create Notification Object                     │
│              {                                              │
│                id: uuid(),                                  │
│                type: 'info' | 'success' | 'warning',       │
│                title: string,                               │
│                message: string,                             │
│                timestamp: Date,                             │
│                read: false,                                 │
│                actionUrl?: string                           │
│              }                                              │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Add to Notifications Array                     │
│              setNotifications((prev) => [...prev, notif])   │
└─────────────────────────────────────────────────────────────┘
                          │
              ┌───────────┴───────────┐
              │                       │
              ▼                       ▼
    ┌──────────────────┐    ┌──────────────────┐
    │  Update Badge    │    │  Show Toast      │
    │  Count           │    │  (if urgent)     │
    └──────────────────┘    └──────────────────┘
              │                       │
              └───────────┬───────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              User Interacts                                 │
└─────────────────────────────────────────────────────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Click       │ │  Mark as     │ │  Dismiss     │
    │  Notification│ │  Read        │ │              │
    └──────────────┘ └──────────────┘ └──────────────┘
              │           │           │
              ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Navigate to │ │  Update read │ │  Remove from │
    │  actionUrl   │ │  status      │ │  array       │
    └──────────────┘ └──────────────┘ └──────────────┘
                          │
                          ▼
              ┌──────────────────────┐
              │  Update Badge Count  │
              └──────────────────────┘
```

---

## 8. Progress Calculation Data Flow

### Portfolio Progress Tracking
```
┌─────────────────────────────────────────────────────────────┐
│              Compliance Tasks Array                         │
│              complianceTasks[portfolioId]                   │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Filter & Count Tasks                           │
│              const totalTasks = tasks.length                │
│              const completedTasks = tasks.filter(           │
│                task => task.status === 'completed'          │
│              ).length                                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Calculate Percentage                           │
│              const progress =                               │
│                (completedTasks / totalTasks) * 100          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Update Multiple UI Elements                    │
└─────────────────────────────────────────────────────────────┘
              │           │           │           │
              ▼           ▼           ▼           ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  Dashboard   │ │  Portfolio   │ │  Detail Page │ │  Checklist   │
    │  Card        │ │  List        │ │  Header      │ │  Progress    │
    │  Progress    │ │  Item        │ │  Progress    │ │  Bar         │
    └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

---

*Last Updated: January 2026*
*Version: 1.2*
*Platform: AIO Fund Manager - Nexxess Business Advisors*
