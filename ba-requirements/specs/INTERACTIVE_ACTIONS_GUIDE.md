# Interactive Actions Guide

## Complete List of Working Actions in AIO Fund Client Portal

This document catalogs all interactive elements and their functionality across the entire application.

---

## 🔐 Authentication & Login Page

### Guest Access
- **Action**: Click "Explore as Guest" button
- **Result**: Opens Guest Learning Page with limited access to videos and AI Assistant
- **Location**: Left sidebar of login page

### Invited Client Login
- **Action**: Enter invitation code + email + password → Click "Activate Account"
- **Demo Codes**: `demo123` or `invite2024`
- **Result**: Full dashboard access with success toast notification
- **Location**: Invited Client tab (default)

### Existing Client Login
- **Action**: Enter email + password → Click "Sign In"
- **Result**: Full dashboard access with welcome back message
- **Location**: Existing Client tab

### Terms of Service
- **Action**: Click "Terms of Service" link
- **Result**: Opens comprehensive Terms of Service modal with scrollable content
- **Location**: Below Activate Account button
- **Features**: Full legal terms, scroll area, "I Understand" button

### Privacy Policy
- **Action**: Click "Privacy Policy" link
- **Result**: Opens detailed Privacy Policy modal
- **Location**: Below Activate Account button
- **Features**: Complete privacy information, GDPR/CCPA compliance details

### Forgot Password
- **Action**: Click "Forgot password?" link
- **Result**: Opens password reset modal
- **Location**: Existing Client tab, above password field
- **Features**: 
  - Email submission form
  - Validation and loading states
  - Success confirmation screen
  - Instructions for email retrieval
  - Option to try different email
  - Return to login

---

## 👤 Guest Learning Page

### Watch Videos
- **Action**: Click "Watch Video" button or video thumbnail
- **Result**: Opens video modal with playback
- **Location**: Video library grid (4 curated videos)
- **Tracking**: Marks videos as watched with green checkmark

### AI Assistant Chat
- **Action**: Switch to "AI Assistant" tab, type message, press Enter or click Send
- **Result**: Receives AI-generated responses about trust activation
- **Location**: AI Assistant tab
- **Features**: Full chat interface, message history, typing indicators

### Request Consultation
- **Action**: Click "Request Consultation" button (multiple locations)
- **Result**: Opens 3-step consultation request modal
- **Locations**: 
  - Header (top right)
  - CTA cards throughout page
  - AI Assistant section
- **Features**: Multi-step form, date/time picker, validation, submission

### Sign In from Guest Mode
- **Action**: Click "Sign In" button
- **Result**: Returns to login page
- **Location**: Header (top right)

---

## 📊 Dashboard Page (Authenticated)

### KPI Cards Actions

#### Upload Document
- **Action**: Click "Upload Now" on "Documents Pending" KPI card
- **Result**: Opens document upload modal
- **Features**: File type selection, file picker, notes field, validation

#### View Calendar
- **Action**: Click "View Calendar" on "Next Appointment" KPI card
- **Result**: Navigates to Calendar page
- **Shows**: All appointments with calendar view

#### View Marketplace
- **Action**: Click "View Marketplace" on "Funds Approved" KPI card
- **Result**: Navigates to Marketplace page
- **Shows**: Available funds and services

### Trust Progress Card

#### View Specific Trust
- **Action**: Click on any trust card (e.g., "Johnson Family Trust")
- **Result**: Navigates to Trust Detail page for that specific trust
- **Shows**: 
  - Detailed progress tracking
  - Document requirements
  - Timeline
  - Images/documents

#### View All Trusts
- **Action**: Click "View All" button on Trust Progress Card
- **Result**: Navigates to My Trust page → Compliance tab
- **Shows**: Full list of all trusts with filtering

### Recent Documents

#### View Document in Gallery
- **Action**: Click on any document thumbnail (4 thumbnails shown)
- **Result**: Opens image gallery modal with lightbox
- **Features**:
  - Full-screen viewing
  - Navigation arrows (previous/next)
  - Image counter (e.g., "1 / 4")
  - Close button
  - Keyboard navigation (arrow keys, Escape)
  - Smooth transitions

### Recent Messages

#### Open Messaging
- **Action**: Click on any message preview
- **Result**: Navigates to Messaging page with that conversation
- **Shows**: Full messaging interface

### Action Items

#### Upload Missing KYC Form
- **Action**: Click "Upload missing KYC form" action item
- **Result**: Opens upload modal with KYC pre-selected
- **Features**: Same as upload document modal

#### Schedule Compliance Review
- **Action**: Click "Schedule compliance review" action item
- **Result**: Opens appointment scheduling modal
- **Features**: Date/time picker, purpose selection, consultant selection

#### Review Activation Summary
- **Action**: Click "Review activation summary" action item
- **Result**: Opens review modal with activation details
- **Shows**: Summary of all completed and pending steps

---

## 📁 My Trust Page

### Tab Navigation
- **Action**: Click tab (Overview, Documents, Compliance, Timeline)
- **Result**: Switches view to selected tab
- **Maintains**: State across navigation

### Document Upload
- **Action**: Click "Upload Document" button (Documents tab)
- **Result**: Opens upload modal
- **Location**: Top right of Documents tab

### Trust Selection
- **Action**: Click on specific trust in Compliance tab
- **Result**: Navigates to Trust Detail page
- **Shows**: Full trust information

### Timeline Filtering
- **Action**: Use filter dropdown in Timeline tab
- **Result**: Filters timeline events by trust
- **Options**: All trusts or specific trust

---

## 🔍 Trust Detail Page

### Navigation
- **Action**: Click "← Back to My Trust" breadcrumb
- **Result**: Returns to My Trust page (Overview tab)

### View All Activities
- **Action**: Click "View All Activities" button
- **Result**: Navigates to My Trust → Timeline tab filtered for this trust
- **Shows**: All timeline events for this specific trust

### Image Gallery
- **Action**: Click on trust document images
- **Result**: Opens lightbox gallery modal
- **Features**: Same as dashboard gallery

---

## 💬 Messaging Page

### Select Conversation
- **Action**: Click on conversation in sidebar
- **Result**: Loads that conversation's messages
- **Shows**: Full message history

### Send Message
- **Action**: Type message → Press Enter or click Send button
- **Result**: Sends message (simulated)
- **Features**: Typing indicator, timestamp

### Search Messages
- **Action**: Type in search box
- **Result**: Filters conversations in real-time
- **Searches**: Names and message content

---

## 🎓 Learning Center Page

### Watch Videos
- **Action**: Click play button or video card
- **Result**: Opens video modal
- **Tracking**: Progress tracking, completion status

### Download Resources
- **Action**: Click "Download" button on resource cards
- **Result**: Initiates PDF download (simulated)
- **Types**: Guides, checklists, templates

### AI Assistant
- **Action**: Same as guest mode, but with more context
- **Result**: Personalized responses based on account status
- **Features**: Account-specific guidance

### Filter by Category
- **Action**: Click category filter chips
- **Result**: Filters videos by selected category
- **Categories**: All, Getting Started, Trust Basics, Compliance, etc.

---

## 🛒 Marketplace Page

### View Fund Details
- **Action**: Click "View Details" on any fund card
- **Result**: Navigates to Fund Detail page
- **Shows**: Comprehensive fund information, performance, allocation

### Request Consultation (from Marketplace)
- **Action**: Click "Request Consultation" button
- **Result**: Opens consultation request modal
- **Context**: Pre-filled with marketplace context

### Filter Funds
- **Action**: Use category tabs or search
- **Result**: Filters visible funds
- **Options**: All, Real Estate, Cryptocurrency, etc.

---

## 💼 Fund Detail Page

### Navigation
- **Action**: Click "← Back to Marketplace"
- **Result**: Returns to Marketplace page

### Invest/Schedule Call
- **Action**: Click "Schedule Consultation" button
- **Result**: Opens scheduling modal
- **Purpose**: Pre-set to investment discussion

### View Documents
- **Action**: Click on fund prospectus/documents
- **Result**: Opens PDF viewer or download

---

## 📅 Calendar Page

### View Appointment Details
- **Action**: Click on any appointment card
- **Result**: Opens Appointment Detail modal
- **Shows**: 
  - Full appointment information
  - Organizer details
  - Meeting link (if video call)
  - Location (if in-person)

### Schedule New Appointment
- **Action**: Click "New Appointment" button
- **Result**: Opens scheduling modal
- **Features**: Full date/time picker, purpose selection

### Calendar Navigation
- **Action**: Click arrows to change month
- **Result**: Updates calendar view
- **Features**: Current month highlight, date selection

### Appointment Actions (from Detail Modal)
- **Edit**: Click "Edit" → Opens scheduling modal with pre-filled data
- **Cancel**: Click "Cancel Appointment" → Confirms and updates status
- **Reschedule**: Click "Reschedule" → Opens scheduling modal
- **Join Video Call**: Click meeting link → Opens in new tab (if video call)

---

## ⚙️ Settings Page

### Update Profile Information
- **Action**: Edit fields → Click "Save Changes"
- **Result**: Updates profile (simulated) with success toast
- **Fields**: Name, email, phone, address

### Change Password
- **Action**: Enter current/new passwords → Click "Update Password"
- **Result**: Password change confirmation
- **Validation**: Strength indicator, matching check

### Notification Preferences
- **Action**: Toggle switches for different notification types
- **Result**: Saves preferences automatically
- **Types**: Email, SMS, Push for various events

### SMS Settings
- **Action**: Toggle "Enable SMS Notifications"
- **Result**: Shows/hides SMS preference options
- **Options**: Appointments, Documents, Compliance, Messages, Learning

### Language Selection
- **Action**: Select from language dropdown
- **Result**: Updates language preference
- **Options**: English, Spanish, French, German, Chinese

### Time Zone
- **Action**: Select from timezone dropdown
- **Result**: Updates timezone preference
- **Used for**: Appointment scheduling, timestamps

---

## 🔔 Notifications System

### View Notifications
- **Action**: Click bell icon in header
- **Result**: Opens notification dropdown panel
- **Shows**: Recent notifications with icons and timestamps

### Mark as Read
- **Action**: Click "Mark as read" on individual notification
- **Result**: Updates notification status, reduces badge counter
- **Visual**: Grays out read notifications

### Dismiss Notification
- **Action**: Click "X" dismiss button
- **Result**: Removes notification from list
- **Updates**: Badge counter

### Navigate from Notification
- **Action**: Click notification text/title
- **Result**: Navigates to relevant page
- **Examples**:
  - Document notification → My Trust page
  - Message notification → Messaging page
  - Appointment notification → Calendar page

### View All Notifications
- **Action**: Click "View All Notifications" at bottom of panel
- **Result**: Closes panel (future: dedicated notifications page)

---

## 👤 Profile Menu (Header)

### Open Profile Menu
- **Action**: Click avatar/initials in top right
- **Result**: Opens dropdown menu

### Navigate to Profile/Settings
- **Action**: Click "Profile" or "Settings"
- **Result**: Navigates to Settings page

### Support
- **Action**: Click "Support"
- **Result**: Opens support modal or email client (planned)

### Logout
- **Action**: Click "Log out"
- **Result**: 
  - Clears authentication
  - Shows success toast
  - Returns to login page
  - Resets all state

---

## 🎯 Global Modals

### Upload Modal
- **Triggers**: Multiple locations (action items, KPI cards, Trust page)
- **Actions**:
  - Select document type dropdown
  - Click to choose file or drag & drop
  - Add optional notes
  - Click "Upload Document"
- **Validation**: 
  - File type check (PDF, JPG, PNG)
  - File size limit (10MB)
  - Required fields
- **Result**: Success toast or error with guidance

### Schedule Appointment Modal
- **Triggers**: Action items, KPI cards, Calendar page
- **Actions**:
  - Select date from calendar picker
  - Choose time slot
  - Select appointment purpose
  - Choose consultant
  - Add optional notes
  - Click "Schedule Appointment"
- **Features**: 
  - Date validation (no past dates)
  - Available time slots
  - Business hours only
- **Result**: SMS confirmation + success toast

### Review Modal
- **Trigger**: "Review activation summary" action item
- **Shows**: 
  - All completed steps (green checkmarks)
  - Pending steps (progress indicators)
  - Next actions
- **Actions**: 
  - Click individual steps for details
  - Click "Continue" to close

### Video Modal
- **Triggers**: Video thumbnails, play buttons
- **Actions**:
  - Play/pause video
  - Adjust volume
  - Fullscreen mode
  - Close modal
- **Features**: 
  - Progress tracking
  - Mark as complete
  - Auto-close option

### Appointment Detail Modal
- **Trigger**: Click appointment in Calendar
- **Shows**: Complete appointment information
- **Actions**:
  - Edit (opens scheduling modal)
  - Cancel (confirmation required)
  - Reschedule (opens scheduling modal)
  - Join video call (if applicable)
  - Close modal

### Consultation Request Modal
- **Triggers**: Guest page, Marketplace
- **3 Steps**:
  1. Personal Info: Name, email, phone, company
  2. Preferences: Interest area, date/time, message
  3. Review: Summary and submission
- **Actions**:
  - Navigate between steps
  - Edit fields
  - Submit request
- **Result**: Success toast + email confirmation

### Terms of Service Modal
- **Trigger**: Click "Terms of Service" link on login
- **Features**:
  - Scrollable content
  - 13 comprehensive sections
  - "I Understand" button
- **Sections**: Agreement, Service description, Responsibilities, etc.

### Privacy Policy Modal
- **Trigger**: Click "Privacy Policy" link on login
- **Features**:
  - Scrollable content
  - 14 detailed sections
  - GDPR/CCPA compliance info
  - "I Understand" button
- **Sections**: Data collection, Usage, Security, Rights, etc.

### Forgot Password Modal
- **Trigger**: Click "Forgot password?" link
- **2 States**:
  1. Email entry form
  2. Success confirmation
- **Actions**:
  - Enter email
  - Submit
  - Try different email
  - Back to login
- **Features**: Validation, loading state, error handling

### Image Gallery Modal
- **Triggers**: Document thumbnails, trust images
- **Actions**:
  - Navigate with arrow buttons
  - Navigate with keyboard (← →)
  - Close with X or Escape key
  - Click outside to close
- **Features**:
  - Full-screen viewing
  - Image counter
  - Smooth transitions
  - Responsive

---

## ⌨️ Keyboard Shortcuts

### Image Gallery
- **Arrow Left** (←): Previous image
- **Arrow Right** (→): Next image
- **Escape**: Close gallery

### Forms
- **Enter**: Submit form (when focused in text input)
- **Tab**: Navigate between fields

### Modals
- **Escape**: Close modal (most modals)

---

## 📱 Responsive Actions

### Mobile Navigation
- **Action**: Click hamburger menu icon
- **Result**: Opens mobile sidebar
- **Shows**: Full navigation menu

### Touch Gestures
- **Swipe**: Navigate image gallery on mobile
- **Tap**: Same as click for all interactive elements
- **Long press**: Context menus (where applicable)

---

## 🎨 Visual Feedback

### Hover States
- **All Buttons**: Color change on hover
- **Cards**: Shadow and border effects
- **Links**: Underline and color change
- **Images**: Scale and overlay effects

### Active States
- **Selected Tab**: Underline and color accent
- **Active Route**: Sidebar highlight
- **Form Focus**: Border color change
- **Checkbox/Toggle**: Visual state change

### Loading States
- **Button**: "Loading..." text with disabled state
- **Modal**: Spinner or skeleton loader
- **Image**: Placeholder until loaded

### Success/Error States
- **Toast Notifications**: Color-coded (green/red)
- **Form Validation**: Border colors and messages
- **Status Badges**: Color-coded by status

---

## 🔄 State Persistence

### Across Navigation
- **Maintains**: Authentication state, user preferences
- **Clears**: Form inputs (on submit), temporary UI state

### Session Management
- **Current**: In-memory (clears on refresh)
- **Planned**: localStorage persistence, auto-refresh

---

## 🧪 Testing Quick Reference

### Test All Login Flows
1. Guest access → Watch video → Request consultation
2. Invited client → Code: `demo123` → Activate
3. Existing client → Any email/password → Sign in
4. Forgot password → Enter email → Confirm
5. Terms/Privacy → Click links → Read → Close

### Test Dashboard Actions
1. Upload document → Select type → Choose file → Upload
2. Schedule appointment → Pick date/time → Confirm
3. View trust → Click trust card → Navigate
4. Open gallery → Click image → Navigate → Close
5. Send message → Type → Send

### Test All Modals
1. Open each modal from multiple entry points
2. Fill forms and submit
3. Test validation errors
4. Test cancel/close actions
5. Test keyboard shortcuts

### Test Notifications
1. Trigger notification (appointment, upload, etc.)
2. Check badge counter
3. Mark as read
4. Dismiss notification
5. Navigate from notification

---

## 🚀 Performance Optimizations

### Lazy Loading
- Images load on scroll/view
- Modals load content on open
- Video players initialize on play

### Optimistic Updates
- Forms show success immediately
- UI updates before API confirmation
- Rollback on error

### Debouncing
- Search inputs (300ms delay)
- Form validation (on blur)
- API calls (prevent duplicate requests)

---

## 🎯 Accessibility Features

### Keyboard Navigation
- All interactive elements are keyboard accessible
- Logical tab order throughout app
- Focus indicators on all focusable elements

### ARIA Labels
- Screen reader support for all actions
- Descriptive labels for icon-only buttons
- Status announcements for dynamic content

### Color Contrast
- WCAG AA compliant color combinations
- Not relying on color alone for information
- High contrast mode support

---

## 📋 Summary

**Total Interactive Actions**: 100+

**Categories**:
- Authentication: 7 actions
- Navigation: 20+ routes and tabs
- Modals: 10 different modals with multiple actions each
- Forms: 15+ forms with submission and validation
- Notifications: 6 notification types with actions
- Gallery: Full image viewing system
- Calendar: Complete scheduling system
- Messaging: Real-time chat interface
- Settings: 20+ configurable preferences

**All actions are functional and provide appropriate feedback through:**
- Toast notifications (success/error)
- Visual state changes
- Loading indicators
- Confirmation modals
- Navigation changes
- Data updates

Every clickable element in the application has a defined action and provides clear user feedback!
