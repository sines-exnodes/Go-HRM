# Notifications System Guide

## Overview
The AIO Fund Client Dashboard includes a comprehensive notifications system that keeps users informed about important updates, messages, and actions related to their trust activation process.

## Features

### 1. **Notification Badge Counter**
- Red circular badge appears on the bell icon when there are unread notifications
- Displays count up to 9 (shows "9+" for 10 or more unread items)
- Updates in real-time as notifications are read or dismissed

### 2. **Notification Panel**
- Click the bell icon to open a dropdown panel
- Panel shows up to 6 recent notifications
- Scrollable list for viewing all notifications
- Clean, organized layout with clear visual hierarchy

### 3. **Notification Types**
Each notification has a specific type with corresponding icon and color:
- **Success** (Green check): Approvals, confirmations, completed actions
- **Warning** (Orange alert): Pending items requiring attention
- **Message** (Blue chat): New messages from team members
- **Info** (Gray clock): General updates, appointments, reminders

### 4. **Notification Features**

#### Visual Indicators
- **Unread notifications**: Light blue background highlight
- **Blue dot**: Appears next to unread notification titles
- **Icon**: Type-specific colored icon on the left
- **Timestamp**: Shows when the notification was received

#### Interaction
- **Click notification**: Marks as read and navigates to related page
- **Hover**: Shows dismiss (X) button
- **Dismiss**: Remove individual notifications
- **Mark all read**: Button to mark all notifications as read at once

### 5. **Smart Navigation**
- Notifications link directly to relevant pages:
  - Document approvals → Trust page
  - Messages → Messaging page
  - Appointments → Calendar page
  - Fund updates → Marketplace page
  - Learning content → Learning Center

## Current Notifications

The system includes 6 sample notifications:

1. **Document Approved** (5 min ago) - Trust Service Agreement approved
2. **New Message** (1 hour ago) - Message from Sarah Johnson
3. **Pending Document** (2 hours ago) - ID verification needs attention
4. **Upcoming Appointment** (3 hours ago) - Meeting scheduled for tomorrow
5. **Fund Investment Confirmed** (1 day ago) - Tech Innovation Fund processed
6. **New Learning Content** (2 days ago) - New video available

## User Flow

### Opening Notifications
1. Look at the bell icon in the top-right header
2. Red badge shows number of unread items
3. Click bell icon to open notification panel

### Reading Notifications
1. Scroll through the list of notifications
2. Unread items have a light blue background
3. Click on any notification to:
   - Mark it as read
   - Navigate to the related page
   - Close the notification panel

### Managing Notifications
1. **Mark all as read**: Click "Mark all read" button in header
2. **Dismiss individual**: Hover over notification, click X button
3. **View all**: Click "View All Notifications" in footer

### Closing Panel
1. Click bell icon again
2. Click outside the panel
3. Click on a notification (navigates and closes)

## Design System

### Colors
- **Unread background**: Blue-50/30 (light blue tint)
- **Unread indicator**: Blue-600 (bright blue dot)
- **Badge**: Red-600 (alert red)
- **Success icon**: Green-600
- **Warning icon**: Orange-600
- **Message icon**: Blue-600
- **Info icon**: Gray-600

### Typography
- **Title**: 14px, dark navy (#0B1930) for unread, gray-900 for read
- **Description**: 14px, gray-600
- **Timestamp**: 12px, gray-500
- **Header**: Default h3 styling

### Layout
- **Panel width**: 420px
- **Max height**: 400px (scrollable)
- **Padding**: 16px per notification
- **Border radius**: 10px (panel), 50% (badge)

## Technical Implementation

### State Management
```tsx
const [notifications, setNotifications] = useState<Notification[]>(initialNotifications);
const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);
```

### Notification Interface
```tsx
interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'message';
  title: string;
  description: string;
  time: string;
  isRead: boolean;
  route?: Route;
}
```

### Key Functions
- `handleNotificationClick()`: Mark as read and navigate
- `markAllAsRead()`: Mark all notifications as read
- `dismissNotification()`: Remove notification from list
- `getNotificationIcon()`: Return appropriate icon based on type

## Future Enhancements

Potential improvements for future versions:
- Real-time notifications via WebSocket
- Sound/desktop notifications for new items
- Notification preferences/settings
- Notification categories and filters
- Search within notifications
- Archive functionality
- Notification history page
- Push notifications for mobile
- Email digest of unread notifications
- Priority/urgent notification highlighting

## Integration Points

The notification system can be triggered from:
- Document upload completions
- Message receipts
- Calendar event reminders
- Fund transaction confirmations
- Trust status updates
- Learning content releases
- System announcements

## Best Practices

1. **Keep notifications actionable**: Each should have a clear next step
2. **Time-sensitive first**: Show newest notifications at the top
3. **Clear descriptions**: Brief but informative text
4. **Appropriate urgency**: Use warning type only when truly needed
5. **Allow dismissal**: Users should control their notification list
6. **Smart navigation**: Link to the most relevant page
7. **Respect user attention**: Don't overwhelm with too many notifications
